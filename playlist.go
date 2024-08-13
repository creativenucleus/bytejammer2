package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
)

const (
	ORDER_ITERATE = iota
	ORDER_RANDOM
)

type PlaylistItem struct {
	location    string
	author      string
	description string
	code        []byte
}

type IPlaylist interface {
}

type Playlist struct {
	order    int
	items    []PlaylistItem
	previous int
}

func NewPlaylist() *Playlist {
	return &Playlist{
		order:    ORDER_ITERATE,
		items:    []PlaylistItem{},
		previous: -1,
	}
}

func (p *Playlist) length() int {
	return len(p.items)
}

func (p *Playlist) isEmpty() bool {
	return len(p.items) == 0
}

func (p *Playlist) getNext() (*PlaylistItem, error) {
	if len(p.items) == 0 {
		return nil, errors.New("Playlist is empty")
	}

	var iItemToPlay int = 0
	if p.order == ORDER_ITERATE {
		iItemToPlay = (p.previous + 1) % len(p.items)
	} else { // Random
		// #TODO: ensure no repeats
		iItemToPlay = rand.Intn(len(p.items))
	}

	item := p.items[iItemToPlay]
	p.previous = iItemToPlay

	if item.code != nil {
		//		fmt.Printf("Cached: %s\n", item.location)
		return &item, nil
	}

	var code []byte
	var err error
	if strings.HasPrefix(item.location, "http://") || strings.HasPrefix(item.location, "https://") {
		code, err = getCodeFromWeb(item.location)
		if err != nil {
			return nil, err
		}
	} else {
		code, err = getCodeFromFile(item.location)
		if err != nil {
			return nil, err
		}
	}

	p.items[iItemToPlay].code = code
	return &p.items[iItemToPlay], nil
}

func getCodeFromWeb(url string) ([]byte, error) {
	respLua, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer respLua.Body.Close()

	if respLua.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("err write: Status Code = %d", respLua.StatusCode)
	}

	data, err := io.ReadAll(respLua.Body)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getCodeFromFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}
