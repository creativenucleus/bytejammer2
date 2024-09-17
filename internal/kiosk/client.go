package kiosk

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer2/config"
	"github.com/creativenucleus/bytejammer2/internal/controlpanel"
	"github.com/creativenucleus/bytejammer2/internal/files"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/creativenucleus/bytejammer2/internal/websocket"

	term "github.com/nsf/termbox-go"
)

var defaultKioskStarterCode = []byte(
	"-- Welcome to the ByteWall!\n" +
		"-- Please delete this code and play.\n" +
		"--\n" +
		"-- Any issues? Find Violet =)\n" +
		"--\n" +
		"-- Have fun!\n" +
		"-- /jtruk + /VioletRaccoon\n" +
		"\n" +
		"function TIC()\n" +
		"	local t=time()*.001\n" +
		"	for y=0,135 do\n" +
		"		for x=0,239 do\n" +
		"			local dx=120-x\n" +
		"			local dy=68-y\n" +
		"			local d=(dx^2+dy^2)^.5\n" +
		"			local a=math.atan2(dy,dx)\n" +
		"			pix(x,y,8+math.sin(d*.1+a-t)*3)\n" +
		"		end\n" +
		"	end\n" +
		"\n" +
		"	local text=\"ByteWall!\"\n" +
		"	local x=50\n" +
		"	local y=75-math.abs(math.sin(t*3)*30)\n" +
		"	print(text,x+1,y+1,15,false,3)\n" +
		"	print(text,x,y,12,false,3)\n" +
		"end",
)

func RunClient(chUserExitRequest <-chan bool, chKeyPress <-chan term.Key, socketURL url.URL, kioskStarterCodePath string) error {
	kioskClientPath := filepath.Join(config.CONFIG.WorkDir, "kiosk-client-snapshots")
	kioskClientPath, err := filepath.Abs(kioskClientPath)
	if err != nil {
		return err
	}

	err = files.EnsurePathExists(kioskClientPath, 0755)
	if err != nil {
		return err
	}

	codeImportPath := filepath.Join(config.CONFIG.WorkDir, "kiosk-client-import.lua")
	codeExportPath := filepath.Join(config.CONFIG.WorkDir, "kiosk-client-export.lua")
	ticManager, err := tic.NewTicManager(&codeImportPath, &codeExportPath)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine("tic-80-client")
	if err != nil {
		return err
	}

	chMakeSnapshot := make(chan message.MsgDataMakeSnapshot)
	chNewPlayer := make(chan bool)
	// #TODO: error handling?!
	go func() error {
		cp := controlpanel.NewKioskClient(config.CONFIG.ControlPanel.Port, chMakeSnapshot, chNewPlayer)
		return cp.Launch()
	}()

	log.GlobalLog.Log("info", fmt.Sprintf("Kiosk Client: Connecting to: %s", socketURL.String()))
	wsConn, err := websocket.NewWebSocketConnection(socketURL)
	if err != nil {
		return err
	}
	log.GlobalLog.Log("info", fmt.Sprintf("Kiosk Client: Connected to: %s", socketURL.String()))

	kioskStarterCode := defaultKioskStarterCode
	if kioskStarterCodePath != "" {
		kioskStarterCode, err = os.ReadFile(kioskStarterCodePath)
		if err != nil {
			return err
		}
	}

	state := tic.MakeTicStateEditor(kioskStarterCode, 1, 1)
	err = ticManager.SetState(state)
	if err != nil {
		return err
	}

	for {
		select {
		case data := <-chMakeSnapshot:
			sendSnapshot(ticManager, kioskClientPath, data.DisplayName, wsConn)

		case <-chNewPlayer:
			newPlayer(ticManager, kioskStarterCode)

		case <-chUserExitRequest:
			return nil
		}
	}
}

func sendSnapshot(ticManager *tic.TicManager, kioskClientPath string, displayName string, wsConn *websocket.WebSocketConnection) error {
	log.GlobalLog.Log("info", fmt.Sprintf("Sending TIC Snapshot %s", displayName))
	state, err := ticManager.GetState()
	if err != nil {
		return err
	}

	// Save a local snapshot for safety...
	timeNow := time.Now()
	fname := fmt.Sprintf("%s-%s.lua", timeNow.Format("20060102150405"), files.SanitiseFilename(displayName))
	fpath := fmt.Sprintf("%s/%s", kioskClientPath, fname)
	fmt.Printf("Sving: %s", fpath)

	err = os.WriteFile(fpath, state.Code, 0644)
	if err != nil {
		return err
	}

	data, err := json.Marshal(tic.MsgTicSnapshotData{
		DisplayName: displayName,
		Code:        state.Code,
	})
	if err != nil {
		return err
	}

	msg := message.Msg{
		Type: message.MsgTypeTicSnapshot,
		Data: data,
	}

	wsConn.Send(msg)
	log.GlobalLog.Log("info", "TIC Snapshot: Sent (not confimation of receipt)")

	return nil
}

func newPlayer(ticManager *tic.TicManager, starterCode []byte) error {
	log.GlobalLog.Log("info", "Starting New Player")

	state := tic.MakeTicStateEditor(starterCode, 1, 1)
	err := ticManager.SetState(state)
	if err != nil {
		return err
	}

	return nil
}
