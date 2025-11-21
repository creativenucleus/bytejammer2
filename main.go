package main

import (
	"errors"
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
	"github.com/creativenucleus/bytejammer2/internal/studio"
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
	// Run keyboard capture in a goroutine.
	// Functions can listen to this on a channel...
	keyboard := keyboard.NewKeyboard()
	go keyboard.Start()

	app := &cli.App{
		Name:  "ByteJammer",
		Usage: "A Multitool for Socket-Based Livecoding",
		//		Version:  "v2.0.0",
		Compiled: time.Now(), // Does this actually work? It feels like it shouldn't!
		Authors: []*cli.Author{{
			Name: "James Rutherford (jtruk)",
		}},
		Copyright: "(c) James Rutherford",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "config",
				Usage: "Load configuration from the specified file",
				Value: "./config.json",
			},
		},
		Before: func(cCtx *cli.Context) error {
			// Load the configuration file before running any other commands
			configFile := cCtx.String("config")
			return config.LoadGlobal(configFile)
		},
		Commands: []*cli.Command{{
			Name:  "sender",
			Usage: "Starts a basic socket sender. It connects to a websocket and periodically posts the contents of a watched file",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "socketurl",
					Usage:    "URL (e.g. ws://drone.alkama.com:9000/room/username)",
					Required: true,
				},
				&cli.StringFlag{
					Name:     "filepath",
					Usage:    "File to watch (e.g. C:/Users/username/Documents/MyFile.lua)",
					Required: true,
				},
			},
			Action: func(cCtx *cli.Context) error {
				inputSocketURL := cCtx.String("socketurl")
				if inputSocketURL == "" {
					panic("socketurl is required")
				}

				socketURL, err := url.Parse(inputSocketURL)
				if err != nil {
					panic(err)
				}
				filepath := cCtx.String("filepath")

				checkFrequency := 2 * time.Second
				return runSender(keyboard.ChUserExitRequest, *socketURL, filepath, checkFrequency)
			},
		}, {
			Name:  "jukebox",
			Usage: "Starts a local jukebox of effects",
			Action: func(*cli.Context) error {
				return runJukebox(keyboard.ChUserExitRequest)
			},
		}, {
			/*			Name:  "bytejam-overlay",
						Usage: "Starts a the overlay - for use with OBS (this is a bit hacky!)",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "sourcefile",
								Usage:    "File to read (e.g. C:/Users/username/Documents/ticws-output.dat)",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "destfile",
								Usage:    "File to output (e.g. C:/Users/username/Documents/MyFile.dat)",
								Required: true,
							},
							&cli.UintFlag{
								Name:     "port",
								Usage:    "Port to display overlay (e.g. 9123 will provide http://localhost:9123)",
								Required: true,
							},
							&cli.StringFlag{
								Name:  "playername",
								Usage: "The player name to display on the overlay",
							},
						},
						Action: func(cCtx *cli.Context) error {
							sourceFilePath := cCtx.String("sourcefile")
							if sourceFilePath == "" {
								panic("sourcefile is required")
							}

							destFilePath := cCtx.String("destfile")
							if destFilePath == "" {
								panic("destfile is required")
							}

							port := cCtx.Uint("port")

							config := controlpanel.ObsOverlayServerConfig{
								ProxyDestFile:  destFilePath,
								PlayerName:     cCtx.String("playername"),
								ObsOverlayPort: port,
							}

							log.GlobalLog.Log("info", fmt.Sprintf("Starting a file proxy (source: %s) (dest: %s)", sourceFilePath, destFilePath))
							chFileUpdated := filewatcher.NewFileWatcher(destFilePath, 100*time.Millisecond, keyboard.ChUserExitRequest)

							return controlpanel.ObsOverlayRun(keyboard.ChUserExitRequest, config, chFileUpdated)
						},
			*/
		}, {
			Name:  "client",
			Usage: "run a default TIC-80 client",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "socketurl",
					Usage:    "URL (e.g. ws://drone.alkama.com:9000/room/username)",
					Required: true,
				},
			},
			Action: func(cCtx *cli.Context) error {
				inputSocketURL := cCtx.String("socketurl")
				if inputSocketURL == "" {
					panic("socketurl is required")
				}

				socketURL, err := url.Parse(inputSocketURL)
				if err != nil {
					panic(err)
				}

				return runClient(keyboard.ChUserExitRequest, *socketURL)
			},
		}, {
			Name:  "server",
			Usage: "Starts a server for managing a livecoding session",
			Action: func(*cli.Context) error {
				return runServer(keyboard.ChUserExitRequest)
			},
		}, {
			Name:  "record",
			Usage: "Starts a recorder that listens to a TIC-80 machine and records snapshots of the machine state",
			Action: func(*cli.Context) error {
				updateDuration := 1 * time.Second
				return runRecorder(keyboard.ChUserExitRequest, updateDuration)
			},
		}, {
			Name:  "replay",
			Usage: "Starts a replayer that re-runs the snapshots from a recorder",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "socketurl",
					Usage:    "URL (e.g. ws://drone.alkama.com:9000/room/username)",
					Required: true,
				},
			},
			Action: func(cCtx *cli.Context) error {
				inputSocketURL := cCtx.String("socketurl")
				if inputSocketURL == "" {
					panic("socketurl is required")
				}

				socketURL, err := url.Parse(inputSocketURL)
				if err != nil {
					panic(err)
				}

				return runReplayer(keyboard.ChUserExitRequest, *socketURL)
			},
		}, {
			Name:  "kiosk-client",
			Usage: "Starts a kiosk client - for players on a ByteWall",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "url",
					Usage:    "URL (e.g. ws://drone.alkama.com:9000/bytejammer/test)",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "startercodepath",
					Usage: "A filepath pointing to a Lua file to be the starter code",
				},
			},
			Action: func(cCtx *cli.Context) error {
				inputURL := cCtx.String("url")
				if inputURL == "" {
					panic("URL is required")
				}

				u, err := url.Parse(inputURL)
				if err != nil {
					panic(err)
				}

				startercodepath := cCtx.String("startercodepath")

				return kiosk.RunClient(keyboard.ChUserExitRequest, keyboard.ChKeyPress, *u, startercodepath)
			},
		}, {
			Name:  "kiosk-server",
			Usage: "Starts a kiosk server - for displaying ByteWall submissions",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "connection",
					Usage:    "One of: client, host",
					Required: true,
				},
				&cli.StringFlag{
					Name:  "url",
					Usage: "URL (e.g. ws://drone.alkama.com:9000/bytejammer/test)",
				},
				&cli.StringFlag{
					Name:  "port",
					Usage: "A port to serve from (e.g. 9123)",
				},
				&cli.StringFlag{
					Name:  "endpoint",
					Usage: "Endpoint to serve from (e.g. /mykiosk/server)",
				},
				&cli.IntFlag{
					Name:  "obs-overlay-port",
					Usage: "Set this if you'd like a web-based OBS overlay",
				},
			},
			Action: func(cCtx *cli.Context) error {
				connectionType := cCtx.String("connection")

				config := kiosk.ServerConfig{}
				switch connectionType {
				case "client":
					inputURL := cCtx.String("url")
					if inputURL == "" {
						panic("URL is required for client connection type")
					}

					clientURL, err := url.Parse(inputURL)
					if err != nil {
						panic(err)
					}

					config.Client.Url = *clientURL
				case "host":
					port := cCtx.Int("port")
					if port == 0 {
						panic("port is required for host connection type")
					}

					endpoint := cCtx.String("endpoint")
					if endpoint == "" {
						panic("endpoint is required for host connection type")
					}

					config.Host.Port = port
					config.Host.Endpoint = endpoint
				default:
					panic("Invalid connection type")
				}

				config.ObsOverlayPort = cCtx.Uint("obs-overlay-port")

				return kiosk.RunServer(keyboard.ChUserExitRequest, config)
			},
		}, {
			// Experimental
			Name:  "studio",
			Usage: "Starts a control panel for the studio - currently just for managing ByteJam server/overlays",
			Flags: []cli.Flag{
				&cli.UintFlag{
					Name:  "port",
					Usage: "A port to serve from (e.g. 9123)",
				},
			},
			Action: func(cCtx *cli.Context) error {
				port := cCtx.Uint("port")
				// TODO check port is non-zero
				if port == 0 {
					return errors.New("port is required")
				}

				return runStudio(keyboard.ChUserExitRequest, port)
			},
		}},
	}

	return app.Run(os.Args)
}

func runSender(chUserExitRequest <-chan bool, socketURL url.URL, filepath string, checkFrequency time.Duration) error {
	// Open a socket
	log.GlobalLog.Log("info", fmt.Sprintf("Sender: Connecting to: %s", socketURL.String()))
	wsConn, err := websocket.NewWebSocketConnection(socketURL)
	if err != nil {
		return err
	}
	log.GlobalLog.Log("info", fmt.Sprintf("Sender: Connected to: %s", socketURL.String()))

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
			case <-chUserExitRequest:
				return
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

func runClient(chUserExitRequest <-chan bool, socketURL url.URL) error {
	codeExportPath := filepath.Join(config.CONFIG.WorkDir, "export.lua")
	ticManager, err := tic.NewTicManager(nil, &codeExportPath)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine("tic-80-client")
	if err != nil {
		return err
	}

	// Open a socket
	log.GlobalLog.Log("info", fmt.Sprintf("Client: Connecting to: %s", socketURL.String()))
	wsClient, err := websocket.NewWebSocketRawDataClient(socketURL)
	if err != nil {
		return err
	}
	ticManager.AddReceiver(wsClient)

	for {
		select {
		case <-chUserExitRequest:
			return nil
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func runServer(chUserExitRequest <-chan bool) error {
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

	// #TODO: Temporary...
	socketURL := url.URL{
		Scheme: "ws",
		Host:   "drone.alkama.com:9000",
		Path:   "/jtruk/test",
	}

	wsClient, err := websocket.NewWebSocketRawDataClient(socketURL)
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
		select {
		case <-chUserExitRequest:
			return nil
		default:
			time.Sleep(1 * time.Second)
		}
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

	chRestartJukebox := make(<-chan bool) // Unused
	LcdzJukebox := jukebox.NewJukebox(playlist)
	LcdzJukebox.AddReceiver(ticManager)
	go func() {
		LcdzJukebox.Run(chRestartJukebox)
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

func runReplayer(chUserExitRequest <-chan bool, socketURL url.URL) error {
	log.GlobalLog.Log("info", "Replayer starting...")

	replayPath := filepath.Join(config.CONFIG.WorkDir, "snaps.zip")
	replayer, err := tic.NewReplayer(replayPath)
	if err != nil {
		return err
	}

	wsClient, err := websocket.NewWebSocketRawDataClient(socketURL)
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

func runStudio(chUserExitRequest <-chan bool, port uint) error {
	// #TODO: error handling?!
	go func() error {
		studio := studio.NewStudio(chUserExitRequest, port)
		return studio.Run()
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
