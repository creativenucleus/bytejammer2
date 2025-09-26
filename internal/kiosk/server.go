package kiosk

import (
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer2/config"
	"github.com/creativenucleus/bytejammer2/internal/controlpanel"
	"github.com/creativenucleus/bytejammer2/internal/files"
	"github.com/creativenucleus/bytejammer2/internal/jukebox"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/playlist"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
)

// should either have a client or a host set
type ServerConfig struct {
	Client struct {
		Url url.URL
	}
	Host struct {
		Port     int
		Endpoint string // starts with /
	}
	ObsOverlayPort uint // optional, if set to non-zero, will run an OBS overlay
}

func RunServer(chUserExitRequest <-chan bool, c ServerConfig) error {
	/*
		_, err := controlpanel.Start(config.CONFIG.ControlPanel.Port)
		if err != nil {
			return err
		}
	*/

	var obsOverlay *controlpanel.ObsOverlayKiosk
	if c.ObsOverlayPort != 0 {
		// #TODO: error handling?!
		go func() error {
			obsOverlay = controlpanel.NewObsOverlayKiosk(c.ObsOverlayPort)
			return obsOverlay.Launch()
		}()
	}

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

	// #TODO: There will be a better way to do this
	if c.Client.Url.String() != "" {
		log.GlobalLog.Log("info", fmt.Sprintf("Kiosk Server: Client for: %s", c.Client.Url.String()))

		wsClient, err := websocket.NewWebSocketConnection(c.Client.Url)
		if err != nil {
			return err
		}
		wsClient.AddReceiver(kioskServer)
	} else if c.Host.Port != 0 || c.Host.Endpoint == "" {
		log.GlobalLog.Log("info", fmt.Sprintf("Kiosk Server: Host at: %d:%s", c.Host.Port, c.Host.Endpoint))

		wsServer, err := websocket.NewWebSocketServer(c.Host.Port, c.Host.Endpoint)
		if err != nil {
			return err
		}
		wsServer.AddReceiver(kioskServer)
	} else {
		return fmt.Errorf("no client or host set properly")
	}

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

	chRestartJukebox := make(chan bool)
	jukebox := jukebox.NewJukebox(playlist)
	jukebox.AddReceiver(ticManager)

	if obsOverlay != nil {
		jukebox.SetObsOverlay(obsOverlay)
	}

	log.GlobalLog.Log("info", fmt.Sprintf("jukebox running from path: %s", kioskPath))
	go func() {
		jukebox.Run(chRestartJukebox)
	}()

	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			// #TODO: Beware of sync issues (e.g. deleted files)
			// #TODO: playlist copy
			isUpdated, err := playlist.SyncWithDirectory(kioskPath)
			if err != nil {
				// #TODO: Error?
				continue
			}

			if isUpdated {
				length := jukebox.Playlist().Length()
				fmt.Printf("Playlist updated - length is now: %d\n", length)
				chRestartJukebox <- true
			}

		case <-chUserExitRequest:
			return nil
		}
	}
}
