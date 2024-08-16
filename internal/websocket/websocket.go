package websocket

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var WsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true }, // #TODO: Check
}

type WebSocket struct {
	Conn *websocket.Conn
}

type HttpHandler func(w http.ResponseWriter, r *http.Request)

type ReadHandler func(WebSocket)

func NewWebSocketHandler(readFn ReadHandler) HttpHandler {
	ws := WebSocket{}

	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		ws.Conn, err = WsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		defer ws.Conn.Close()

		// #TODO: handle exit
		for {
			readFn(ws)
		}
	}
}
