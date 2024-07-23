package main

import (
	"fmt"
	"time"
)

type Jukebox struct {
	MessageBroadcaster
	playlist Playlist
}

func NewJukebox(playlist Playlist) *Jukebox {
	l := &Jukebox{
		playlist: playlist,
	}
	return l
}

func (j *Jukebox) run() {
	for {
		item, err := j.playlist.getNext()
		if err != nil {
			panic(err) // #TODO!
		}

		//		state := MakeTicStateEditor([]byte("Hey there, how are you\nocwenocwen\nocwenowcniciowe"), 4, 2)
		state := MakeTicStateRunning(item.code)
		//state := MakeTicStateEditor([]byte(item.code), 4, 2)
		fmt.Printf("Sending %s %s %s\n", item.author, item.description, item.location)

		m := NewMessageTicState(state)
		j.broadcast(&m)
		time.Sleep(10 * time.Second)
	}
}
