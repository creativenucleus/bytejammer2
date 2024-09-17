package kiosk

import (
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer2/config"
	"github.com/creativenucleus/bytejammer2/internal/files"
	"github.com/creativenucleus/bytejammer2/internal/jukebox"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/playlist"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
)

func RunServer(chUserExitRequest <-chan bool, socketURL url.URL) error {
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

	playlist, err := playlist.NewPlaylistDirectory(kioskPath)
	if err != nil {
		return err
	}

	jukebox := jukebox.NewJukebox(*playlist)
	jukebox.AddReceiver(ticManager)
	log.GlobalLog.Log("info", fmt.Sprintf("jukebox running from path: %s", kioskPath))
	go func() {
		jukebox.Run()
	}()

	ticker := time.NewTicker(5 * time.Second)
	lastPlaylistLength := -1
	for {
		select {
		case <-ticker.C:
			// #TODO: Beware of sync issues (e.g. deleted files)
			// #TODO: playlist copy
			length := jukebox.Playlist().Length()
			playlist := jukebox.Playlist()
			err := playlist.SyncWithDirectory(kioskPath)
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
