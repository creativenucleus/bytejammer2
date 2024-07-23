package main

import (
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer2/config"
)

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

func run() error {
	err := config.Load("./config.json")
	if err != nil {
		return err
	}

	codeImportPath := filepath.Join(config.CONFIG.WorkDir, "import.lua")
	//	codeExportPath := filepath.Join(config.CONFIG.WorkDir, "export.lua")

	tic, err := NewTic(&codeImportPath, nil) //codeExportPath)
	if err != nil {
		return err
	}

	err = tic.startMachine()
	if err != nil {
		return err
	}

	playlist, err := NewPlaylistLCDZ()
	if err != nil {
		panic(err) // #TODO!
	}

	LcdzJukebox := NewJukebox(*playlist)
	LcdzJukebox.addReceiver(tic)
	go func() {
		LcdzJukebox.run()
	}()

	/*
		_, err = controlpanel.Start(8080)
		if err != nil {
			return err
		}
	*/

	/*
		wsClient, err := NewWebSocketClient("host", 8080, "/path")
		if err != nil {
			panic(err)
		}
		wsClient.addReceiver(tic)
	*/

	for {
		time.Sleep(1 * time.Second)
	}
}
