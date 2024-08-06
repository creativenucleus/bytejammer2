package main

import (
	"os"
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer2/config"
	"github.com/creativenucleus/bytejammer2/internal/controlpanel"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
	"github.com/urfave/cli/v2"
)

func main() {
	err := runCli()
	if err != nil {
		log.GlobalLog.Send(&message.Msg{Type: message.MsgTypeLog, Data: log.MsgLogData{Level: "error", Message: err.Error()}})
	}
}

func runCli() error {
	err := config.LoadGlobal("./config.json")
	if err != nil {
		return err
	}

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
			{
				Name:  "jukebox",
				Usage: "run a default jukebox",
				Action: func(*cli.Context) error {
					return runJukebox()
				},
			},
		},
	}

	return app.Run(os.Args)
}

func runClient() error {
	codeExportPath := filepath.Join(config.CONFIG.WorkDir, "export.lua")
	ticManager, err := tic.NewTicManager(nil, &codeExportPath)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine()
	if err != nil {
		return err
	}

	wsClient, err := websocket.NewWebSocketClient("drone.alkama.com", 9000, "/bytejammer/test")
	if err != nil {
		return err
	}
	ticManager.AddReceiver(wsClient)

	for {
		time.Sleep(1 * time.Second)
	}
}

func runServer() error {
	_, err := controlpanel.Start(config.CONFIG.ControlPanel.Port)
	if err != nil {
		return err
	}

	for {
		time.Sleep(1 * time.Second)
	}
}

func runJukebox() error {
	codeImportPath := filepath.Join(config.CONFIG.WorkDir, "import.lua")
	ticManager, err := tic.NewTicManager(&codeImportPath, nil)
	if err != nil {
		return err
	}

	err = ticManager.StartMachine()
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
		time.Sleep(1 * time.Second)
	}
}
