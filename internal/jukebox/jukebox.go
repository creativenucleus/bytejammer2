package jukebox

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/creativenucleus/bytejammer2/config"
	"github.com/creativenucleus/bytejammer2/internal/controlpanel"
	"github.com/creativenucleus/bytejammer2/internal/files"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/playlist"
	"github.com/creativenucleus/bytejammer2/internal/tic"
)

// Jukebox contains a playlist and periodically sends a TIC-80 state from the playlist
type Jukebox struct {
	message.MsgPropagator
	playlist      *playlist.Playlist
	sceneDuration time.Duration
	obsOverlay    *controlpanel.ObsOverlayKiosk // Optional OBS overlay
}

func NewJukebox(playlist *playlist.Playlist) *Jukebox {
	l := &Jukebox{
		playlist:      playlist,
		sceneDuration: time.Duration(uint(time.Second) * config.CONFIG.Jukebox.RotatePeriodInSeconds),
	}
	return l
}

func (j *Jukebox) SetObsOverlay(obsOverlay *controlpanel.ObsOverlayKiosk) {
	j.obsOverlay = obsOverlay
}

func (j *Jukebox) SetSceneDuration(sceneDuration time.Duration) {
	j.sceneDuration = sceneDuration
}

func (j *Jukebox) Playlist() *playlist.Playlist {
	return j.playlist
}

func (j *Jukebox) Run(forceRestart <-chan bool) error {
	log.GlobalLog.Log("info", "Jukebox starting with rotation period of "+j.sceneDuration.String())

	sceneTicker := time.NewTicker(j.sceneDuration)

	for {
		select {
		case <-sceneTicker.C:
			j.playNext()
		case <-forceRestart:
			j.playNext()
			sceneTicker.Reset(j.sceneDuration)
		}
	}
}

func (j *Jukebox) playNext() error {
	item, err := j.playlist.GetNext()
	if err != nil {
		// TODO: Handle error / log
		return err
	}

	state := tic.MakeTicStateRunning(item.Code())
	log.GlobalLog.Log("info", fmt.Sprintf("Jukebox running Lua file: %s %s %s", item.Author(), item.Description(), item.Location()))

	data, err := json.Marshal(state)
	if err != nil {
		return err
	}

	j.Propagate(message.MsgTypeTicState, data)

	// Try loading metadata for the item...
	filePathMetaJson := fmt.Sprintf("%s.meta.json", item.Location())
	metaJson, err := files.LoadMetaJson(filePathMetaJson)
	playerName := ""
	effectName := ""
	if err != nil {
		// Ignore 'file doesn't exist' error - it might legitimately not be there
		if !os.IsNotExist(err) {
			return err
		}
	} else {
		playerName = metaJson["player_name"]
		effectName = metaJson["effect_name"]
	}

	if j.obsOverlay != nil {
		err = j.obsOverlay.SetDetail(playerName, effectName)
		if err != nil {
			log.GlobalLog.Log("error", fmt.Sprintf("Failed to set OBS overlay details: %v", err))
		}
	}

	return nil
}
