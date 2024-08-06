package main

import (
	"fmt"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/tic"
)

const defaultJukeboxSceneDuration = time.Duration(10 * time.Second)

// Jukebox contains a playlist and periodically sends a TIC-80 state from the playlist
type Jukebox struct {
	message.MsgSender
	playlist      Playlist
	sceneDuration time.Duration
}

func NewJukebox(playlist Playlist) *Jukebox {
	l := &Jukebox{
		playlist:      playlist,
		sceneDuration: defaultJukeboxSceneDuration,
	}
	return l
}

func (j *Jukebox) SetSceneDuration(sceneDuration time.Duration) {
	j.sceneDuration = sceneDuration
}

func (j *Jukebox) run() error {
	for {
		item, err := j.playlist.getNext()
		if err != nil {
			return err
		}

		state := tic.MakeTicStateRunning(item.code)
		log.GlobalLog.Send(&message.Msg{Type: message.MsgTypeLog, Data: log.MsgLogData{
			Level:   "info",
			Message: fmt.Sprintf("Sending %s %s %s\n", item.author, item.description, item.location),
		}})

		msg := tic.NewMessageTicState(state)
		j.Send(msg)
		time.Sleep(j.sceneDuration)
	}
}
