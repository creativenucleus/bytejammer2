package main

import (
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

// Try to join host:port/path with a websocket connection
func NewWebSocketClient(host string, port int, path string) (*WebSocket, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   path,
	}
	log.Printf("-> Connecting to %s", u.String())

	s := WebSocket{}
	var err error
	s.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return &s, nil
}
