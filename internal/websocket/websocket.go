package websocket

import (
	"fmt"
	"log"
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

	fmt.Println("NewWebSocketHandler->added")
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		fmt.Println("NewWebSocketHandler->upgrade")
		ws.Conn, err = WsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		fmt.Println("NewWebSocketHandler->upgraded")

		defer ws.Conn.Close()

		fmt.Println("NewWebSocketHandler->loop")

		// #TODO: handle exit
		for {
			readFn(ws)
		}
	}
}
