package kiosk

import (
	_ "embed"
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

//go:embed defaultKioskStarterCode.lua
var defaultKioskStarterCode []byte

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
			sendSnapshot(ticManager, kioskClientPath, data.PlayerName, data.EffectName, wsConn)

		case <-chNewPlayer:
			newPlayer(ticManager, kioskStarterCode)

		case <-chUserExitRequest:
			return nil
		}
	}
}

func sendSnapshot(
	ticManager *tic.TicManager,
	kioskClientPath string,
	playerName string,
	effectName string,
	wsConn *websocket.WebSocketConnection,
) error {
	log.GlobalLog.Log("info", fmt.Sprintf("Sending TIC Snapshot %s (%s)", playerName, effectName))
	state, err := ticManager.GetState()
	if err != nil {
		return err
	}

	// Save a local snapshot for safety...
	timeNow := time.Now()
	fnameBase := fmt.Sprintf("%s-%s-%s", timeNow.Format("20060102150405"), files.SanitiseFilename(playerName), files.SanitiseFilename(effectName))
	fpathLua := fmt.Sprintf("%s/%s.lua", kioskClientPath, fnameBase)
	fpathMetaJson := fmt.Sprintf("%s.meta.json", fpathLua)

	fmt.Printf("Saving: %s", fpathLua)
	err = os.WriteFile(fpathLua, state.Code, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Saving: %s", fpathMetaJson)
	err = files.SaveMetaJson(fpathMetaJson, playerName, effectName)
	if err != nil {
		return err
	}

	data, err := json.Marshal(tic.MsgTicSnapshotData{
		PlayerName: playerName,
		EffectName: effectName,
		Code:       state.Code,
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
