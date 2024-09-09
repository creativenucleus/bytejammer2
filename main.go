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
			{
				Name:  "sender",
				Usage: "run a socket sender",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "socketurl",
						Usage:    "URL (e.g. ws://drone.alkama.com:9000/room/username)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "file",
						Usage:    "File to watch (e.g. C:/Users/username/Documents/MyFile.lua)",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					socketURL, err := url.Parse(cCtx.String("socketurl"))
					if err != nil {
						panic(err)
					}
					filepath := cCtx.String("filepath")

					checkFrequency := 2 * time.Second
					return runSender(*socketURL, filepath, checkFrequency)
				},
			},
			/*			{
							Name:  "client",
							Usage: "run a default client",
							Action: func(*cli.Context) error {
								return runClient()
							},
						},
			*/

			{
				Name:  "server",
				Usage: "run a default server",
				Action: func(*cli.Context) error {
					return runServer()
				},
			},
			{
				Name:  "kiosk-client",
				Usage: "run a kiosk client",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "socketurl",
						Usage:    "URL (e.g. ws://drone.alkama.com:9000/bytejammer/evoke)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "startercodepath",
						Usage: "A filepath pointing to a Lua file to be the starter code",
					},
				},
				Action: func(cCtx *cli.Context) error {
					u, err := url.Parse(cCtx.String("socketurl"))
					if err != nil {
						panic(err)
					}

					startercodepath := cCtx.String("startercodepath")

					return runKioskClient(keyboard.ChUserExitRequest, keyboard.ChKeyPress, *u, startercodepath)
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
			{
				Name:  "recorder",
				Usage: "run a default recorder",
				Action: func(*cli.Context) error {
					return runRecorder(keyboard.ChUserExitRequest)
				},
			},
			{
				Name:  "replayer",
				Usage: "run a default replayer",
				Action: func(*cli.Context) error {
					return runReplayer(keyboard.ChUserExitRequest)
				},
			},
		},
	}

	return app.Run(os.Args)
}

func runSender(socketURL url.URL, filepath string, checkFrequency time.Duration) error {
	// Open a socket
	log.GlobalLog.Log("info", fmt.Sprintf("Kiosk Client: Connecting to: %s", socketURL.String()))
	wsConn, err := websocket.NewWebSocketConnection(socketURL)
	if err != nil {
		return err
	}
	log.GlobalLog.Log("info", fmt.Sprintf("Kiosk Client: Connected to: %s", socketURL.String()))

	// Watch a file
	// Check frequency
	// Only send if file has changed
	// Send frequency
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)

	chFileDataUpdate := make(chan []byte)
	filewatcher, err := files.NewFileWatcher(filepath, checkFrequency, chFileDataUpdate)
	if err != nil {
		return err
	}

	dataToSend, err := filewatcher.Read()
	if err != nil {
		return err
	}

	chError := make(chan error)
	// TODO: something with chError

	go func() {
		for {
			select {
			case <-done:
				return
			case dataFromFile := <-chFileDataUpdate:
				dataToSend = dataFromFile
				err = sendRawData(wsConn, dataToSend)
				if err != nil {
					chError <- err
				}

			case <-ticker.C:
				dataToSend, err := filewatcher.Read()
				if err != nil {
					chError <- err
				}
				err = sendRawData(wsConn, dataToSend)
				if err != nil {
					chError <- err
				}
			}
		}
	}()

	go func() {
		for {
			filewatcher.Run()
		}
	}()

	time.Sleep(1600 * time.Millisecond)
	ticker.Stop()
	done <- true

	fmt.Println("Ticker stopped")

	return nil
}

func runKioskClient(chUserExitRequest <-chan bool, chKeyPress <-chan term.Key, socketURL url.URL, kioskStarterCodePath string) error {
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

func sendRawData(wsConn *websocket.WebSocketConnection, data []byte) error {
	err := wsConn.SendRaw(data)
	if err != nil {
		return err
	}

	log.GlobalLog.Log("info", "Raw Data: Sent (not confimation of receipt)")
	return nil
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

func runKioskServer(chUserExitRequest <-chan bool, socketURL url.URL) error {
	/*
		_, err := controlpanel.Start(config.CONFIG.ControlPanel.Port)
		if err != nil {
			return err
		}
	*/
	kioskPath := filepath.Join(config.CONFIG.WorkDir, "kiosk-server-playlist")
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
	codeImportPath := filepath.Join(config.CONFIG.WorkDir, "kiosk-server-import.lua")
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

	wsClient, err := websocket.NewWebSocketRawDataClient("drone.alkama.com", 9000, "/jtruk/test")
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

	wsClient, err := websocket.NewWebSocketRawDataClient("drone.alkama.com", 9000, "/jtruk/test")
	if err != nil {
		return err
	}
	wsClient.AddReceiver(ticManager)

	// #TODO: error handling?!
	go func() error {
		cp := controlpanel.NewServerPanel(config.CONFIG.ControlPanel.Port)
		return cp.Launch()
	}()

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

func runRecorder(chUserExitRequest <-chan bool) error {
	recorderPath := filepath.Join(config.CONFIG.WorkDir, "recorder")
	err := files.EnsurePathExists(recorderPath, 0755)
	if err != nil {
		return err
	}

	codeExportPath := filepath.Join(config.CONFIG.WorkDir, "export.lua")
	ticManager, err := tic.NewTicManager(nil, &codeExportPath)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine("tic-80-client")
	if err != nil {
		return err
	}

	recorder, err := tic.NewRecorder(recorderPath)
	if err != nil {
		return err
	}
	ticManager.AddReceiver(recorder)

	for {
		select {
		case <-chUserExitRequest:
			return nil
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func runReplayer(chUserExitRequest <-chan bool) error {
	log.GlobalLog.Log("info", "Replayer starting...")

	replayPath := filepath.Join(config.CONFIG.WorkDir, "snaps.zip")
	replayer, err := tic.NewReplayer(replayPath)

	wsClient, err := websocket.NewWebSocketRawDataClient("drone.alkama.com", 9000, "/jtruk/test")
	if err != nil {
		return err
	}
	replayer.AddReceiver(wsClient)

	replayer.Run(chUserExitRequest)
	/*
		for {
			select {
			case <-chUserExitRequest:
				return nil
			default:
				time.Sleep(1 * time.Second)
			}
		}*/
	return nil
}
