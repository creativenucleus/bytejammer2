package main

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // #TODO: Check
}

type HttpHandler func(w http.ResponseWriter, r *http.Request)

func NewWebSocketHandler(host string, port int, path string) (HttpHandler, error) {
	ws := WebSocket{}

	wsHandler := func(w http.ResponseWriter, r *http.Request) {
		var err error
		ws.conn, err = wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		//		defer cp.wsClient.Close()

		go ws.listen()

		// #TODO: handle exit
		for {
		}
	}

	return wsHandler, nil
}
