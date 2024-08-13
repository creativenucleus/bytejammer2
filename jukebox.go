package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/creativenucleus/bytejammer2/config"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/tic"
)

// Jukebox contains a playlist and periodically sends a TIC-80 state from the playlist
type Jukebox struct {
	message.MsgPropagator
	playlist      Playlist
	sceneDuration time.Duration
}

func NewJukebox(playlist Playlist) *Jukebox {
	l := &Jukebox{
		playlist:      playlist,
		sceneDuration: time.Duration(uint(time.Second) * config.CONFIG.Jukebox.RotatePeriodInSeconds),
	}
	return l
}

func (j *Jukebox) SetSceneDuration(sceneDuration time.Duration) {
	j.sceneDuration = sceneDuration
}

func (j *Jukebox) run() error {
	log.GlobalLog.Log("info", "Jukebox starting with rotation period of "+j.sceneDuration.String())

	for {
		item, err := j.playlist.getNext()
		if err != nil {
			// TODO: Handle error / log
			continue
		}

		state := tic.MakeTicStateRunning(item.code)
		log.GlobalLog.Log("info", fmt.Sprintf("Jukebox running Lua file: %s %s %s", item.author, item.description, item.location))

		data, err := json.Marshal(state)
		if err != nil {
			return err
		}

		j.Propagate(message.MsgTypeTicState, data)
		time.Sleep(j.sceneDuration)
	}
}
