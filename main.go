package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer2/config"
	"github.com/creativenucleus/bytejammer2/internal/controlpanel"
	"github.com/creativenucleus/bytejammer2/internal/files"
	"github.com/creativenucleus/bytejammer2/internal/jukebox"
	"github.com/creativenucleus/bytejammer2/internal/keyboard"
	"github.com/creativenucleus/bytejammer2/internal/kiosk"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/playlist"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
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
						Usage:    "URL (e.g. ws://drone.alkama.com:9000/bytejammer/test)",
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

					return kiosk.RunClient(keyboard.ChUserExitRequest, keyboard.ChKeyPress, *u, startercodepath)
				},
			},
			{
				Name:  "kiosk-server",
				Usage: "run a kiosk server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "socketurl",
						Usage:    "URL (e.g. ws://drone.alkama.com:9000/bytejammer/test)",
						Required: true,
					},
				},
				Action: func(cCtx *cli.Context) error {
					u, err := url.Parse(cCtx.String("socketurl"))
					if err != nil {
						panic(err)
					}

					return kiosk.RunServer(keyboard.ChUserExitRequest, *u)
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
					updateDuration := 1 * time.Second
					return runRecorder(keyboard.ChUserExitRequest, updateDuration)
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

func sendRawData(wsConn *websocket.WebSocketConnection, data []byte) error {
	err := wsConn.SendRaw(data)
	if err != nil {
		return err
	}

	log.GlobalLog.Log("info", "Raw Data: Sent (not confimation of receipt)")
	return nil
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

	playlist, err := playlist.NewPlaylistLCDZ()
	if err != nil {
		return err
	}

	LcdzJukebox := jukebox.NewJukebox(*playlist)
	LcdzJukebox.AddReceiver(ticManager)
	go func() {
		LcdzJukebox.Run()
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

func runRecorder(chUserExitRequest <-chan bool, updateDuration time.Duration) error {
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
			return recorder.Close()
		default:
			time.Sleep(updateDuration)
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
