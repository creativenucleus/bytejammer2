package main

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
	"github.com/creativenucleus/bytejammer2/internal/keyboard"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
	term "github.com/nsf/termbox-go"
	"github.com/urfave/cli/v2"
)

func main() {
	err := runCli()
	if err != nil {
		log.GlobalLog.Log("error", err.Error())
	}
}

func runCli() error {
	err := config.LoadGlobal("./config.json")
	if err != nil {
		return err
	}

	keyboard := keyboard.NewKeyboard()
	go keyboard.Start()

	app := &cli.App{
		/*        Flags: []cli.Flag{
		          &cli.StringFlag{
		              Name:    "lang",
		              Aliases: []string{"l"},
		              Value:   "english",
		              Usage:   "Language for the greeting",
		          },
		          &cli.StringFlag{
		              Name:    "config",
		              Aliases: []string{"c"},
		              Usage:   "Load configuration from `FILE`",
		          },
		      },*/
		Commands: []*cli.Command{
			/*			{
							Name:  "client",
							Usage: "run a default client",
							Action: func(*cli.Context) error {
								return runClient()
							},
						},
						{
							Name:  "server",
							Usage: "run a default client",
							Action: func(*cli.Context) error {
								return runServer()
							},
						},
			*/
			{
				Name:  "kiosk-client",
				Usage: "run a kiosk client",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "socketurl",
						Usage:    "URL (e.g. ws://drone.alkama.com:9000/bytejammer/evoke)",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					u, err := url.Parse(cCtx.String("socketurl"))
					if err != nil {
						panic(err)
					}

					return runKioskClient(keyboard.ChUserExitRequest, keyboard.ChKeyPress, *u)
				},
			},
			{
				Name:  "kiosk-server",
				Usage: "run a kiosk server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "socketurl",
						Usage:    "URL (e.g. ws://drone.alkama.com:9000/bytejammer/evoke)",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					u, err := url.Parse(cCtx.String("socketurl"))
					if err != nil {
						panic(err)
					}

					return runKioskServer(keyboard.ChUserExitRequest, *u)
				},
			},
			{
				Name:  "jukebox",
				Usage: "run a default jukebox",
				Action: func(*cli.Context) error {
					return runJukebox(keyboard.ChUserExitRequest)
				},
			},
		},
	}

	return app.Run(os.Args)
}

func runKioskClient(chUserExitRequest <-chan bool, chKeyPress <-chan term.Key, socketURL url.URL) error {
	codeExportPath := filepath.Join(config.CONFIG.WorkDir, "export.lua")
	ticManager, err := tic.NewTicManager(nil, &codeExportPath)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine("tic-80-client")
	if err != nil {
		return err
	}

	chMakeSnapshot := make(chan bool)
	// #TODO: error handling?!
	go func() error {
		cp := controlpanel.NewKioskClient(config.CONFIG.ControlPanel.Port, chMakeSnapshot)
		return cp.Launch()
	}()

	log.GlobalLog.Log("info", fmt.Sprintf("Kiosk Client: Connecting to: %s", socketURL.String()))
	wsConn, err := websocket.NewWebSocketConnection(socketURL)
	if err != nil {
		return err
	}
	log.GlobalLog.Log("info", fmt.Sprintf("Kiosk Client: Connected to: %s", socketURL.String()))

	fmt.Println("Awaiting")
	for {
		select {
		case <-chMakeSnapshot:
			fmt.Println("Yay")
			sendSnapshot(ticManager, wsConn)

		case kp := <-chKeyPress:
			if kp == 32 { // Spacebar
				sendSnapshot(ticManager, wsConn)
			}

		case <-chUserExitRequest:
			return nil
		}
	}
}

func sendSnapshot(ticManager *tic.TicManager, wsConn *websocket.WebSocketConnection) error {
	log.GlobalLog.Log("info", "Sending TIC Snapshot")
	fmt.Println("Sending Snapshot!")
	state, err := ticManager.GetState()
	if err != nil {
		return err
	}

	data, err := json.Marshal(tic.MsgTicSnapshotData{
		ClientID: "kiosk-client",
		Code:     state.Code,
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

func runKioskServer(chUserExitRequest <-chan bool, socketURL url.URL) error {
	/*
		_, err := controlpanel.Start(config.CONFIG.ControlPanel.Port)
		if err != nil {
			return err
		}
	*/
	kioskPath := filepath.Join(config.CONFIG.WorkDir, "kiosk-playlist")
	kioskPath, err := filepath.Abs(kioskPath)
	if err != nil {
		return err
	}

	err = files.EnsurePathExists(kioskPath, 0755)
	if err != nil {
		return err
	}

	// Set up Kiosk Server - this listens for snapshots and adds them to the directory
	kioskServer := tic.NewKioskServer(kioskPath)
	log.GlobalLog.Log("info", fmt.Sprintf("Kiosk Server: Connecting to: %s", socketURL.String()))
	wsConn, err := websocket.NewWebSocketConnection(socketURL)
	if err != nil {
		return err
	}
	wsConn.AddReceiver(kioskServer)

	// Set up the TIC, this picks from the playlist directory...
	codeImportPath := filepath.Join(config.CONFIG.WorkDir, "import.lua")
	ticManager, err := tic.NewTicManager(&codeImportPath, nil)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine("tic-80-server")
	if err != nil {
		return err
	}

	playlist, err := NewPlaylistDirectory(kioskPath)
	if err != nil {
		return err
	}

	jukebox := NewJukebox(*playlist)
	jukebox.AddReceiver(ticManager)
	log.GlobalLog.Log("info", fmt.Sprintf("jukebox running from path: %s", kioskPath))
	go func() {
		jukebox.run()
	}()

	ticker := time.NewTicker(5 * time.Second)
	lastPlaylistLength := -1
	for {
		select {
		case <-ticker.C:
			// #TODO: Beware of sync issues (e.g. deleted files)
			// #TODO: playlist copy
			length := jukebox.playlist.length()
			err := jukebox.playlist.syncWithDirectory(kioskPath)
			if err != nil {
				// #TODO: Error?
				continue
			}
			if length != lastPlaylistLength {
				fmt.Printf("Playlist length is now: %d\n", length)
			}
			lastPlaylistLength = length

		case <-chUserExitRequest:
			return nil
		}
	}
}

func runClient() error {
	codeExportPath := filepath.Join(config.CONFIG.WorkDir, "export.lua")
	ticManager, err := tic.NewTicManager(nil, &codeExportPath)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine("tic-80-client")
	if err != nil {
		return err
	}

	wsClient, err := websocket.NewWebSocketRawDataClient("drone.alkama.com", 9000, "/bytejammer/test")
	if err != nil {
		return err
	}
	ticManager.AddReceiver(wsClient)

	for {
		time.Sleep(1 * time.Second)
	}
}

func runServer() error {
	/*
		_, err := controlpanel.Start(config.CONFIG.ControlPanel.Port)
		if err != nil {
			return err
		}
	*/

	codeImportPath := filepath.Join(config.CONFIG.WorkDir, "import.lua")
	ticManager, err := tic.NewTicManager(&codeImportPath, nil)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine("tic-80-server")
	if err != nil {
		return err
	}

	wsClient, err := websocket.NewWebSocketRawDataClient("drone.alkama.com", 9000, "/bytejammer/test")
	if err != nil {
		return err
	}
	wsClient.AddReceiver(ticManager)

	for {
		time.Sleep(1 * time.Second)
	}
}

func runJukebox(chUserExitRequest <-chan bool) error {
	codeImportPath := filepath.Join(config.CONFIG.WorkDir, "import.lua")
	ticManager, err := tic.NewTicManager(&codeImportPath, nil)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine("tic-80-server")
	if err != nil {
		return err
	}

	playlist, err := NewPlaylistLCDZ()
	if err != nil {
		return err
	}

	LcdzJukebox := NewJukebox(*playlist)
	LcdzJukebox.AddReceiver(ticManager)
	go func() {
		LcdzJukebox.run()
	}()

	for {
		select {
		case <-chUserExitRequest:
			return nil
		default:
			time.Sleep(1 * time.Second)
		}
	}
}
