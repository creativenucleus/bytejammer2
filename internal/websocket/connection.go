package websocket

import (
	"fmt"
	"net/url"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	message.MsgPropagator
	url  url.URL
	conn *websocket.Conn
}

// Try to join host:port/path with a websocket connection
func NewWebSocketConnection(u url.URL) (*WebSocketConnection, error) {
	wsc := &WebSocketConnection{
		url: u,
	}

	err := wsc.dialConnection()
	if err != nil {
		return nil, err
	}

	return wsc, nil
}

func (wsc *WebSocketConnection) dialConnection() error {
	go func() {
		for {
			fmt.Println("Dialing...")
			var err error
			wsc.conn, _, err = websocket.DefaultDialer.Dial(wsc.url.String(), nil)
			if err != nil {
				fmt.Println("Failed to connect... will retry shortly")
				time.Sleep(10 * time.Second)
				continue
			}
			fmt.Println("Connection successful")

			err = wsc.listen()
			wsc.conn = nil
			if err == nil {
				return // break out of the loop if no error
			}
		}
	}()

	return nil
}

func (wsc *WebSocketConnection) Send(msg message.Msg) error {
	if wsc.conn == nil {
		fmt.Println("No active connection - no message sent (stalling)")
		return nil
	}
	return wsc.conn.WriteJSON(msg)
}

func (wsc *WebSocketConnection) SendRaw(data []byte) error {
	if wsc.conn == nil {
		fmt.Println("No active connection - no message sent (stalling)")
		return nil
	}

	return wsc.conn.WriteMessage(websocket.TextMessage, data)
}

func (ws *WebSocketConnection) listen() error {
	return propagateIncomingMessages(ws.conn, func(msgType message.MsgType, msgData []byte) {
		ws.Propagate(msgType, msgData)
	})
}
