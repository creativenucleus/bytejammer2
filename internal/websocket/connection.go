package websocket

import (
	"encoding/json"
	"log"
	"net/url"

	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/gorilla/websocket"
)

type WebSocketConnection struct {
	message.MsgPropagator
	conn *websocket.Conn
}

// Try to join host:port/path with a websocket connection
func NewWebSocketConnection(u url.URL) (*WebSocketConnection, error) {
	ws := &WebSocketConnection{}
	var err error
	ws.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	go ws.listen()

	return ws, nil
}

func (conn *WebSocketConnection) Send(msg message.Msg) error {
	return conn.conn.WriteJSON(msg)
}

func (ws *WebSocketConnection) listen() {
	for {
		messageType, data, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			continue
		}

		if messageType != websocket.TextMessage {
			log.Println("messageType is not Text")
			continue
		}

		var msg message.Msg
		err = json.Unmarshal(data, &msg)
		if err != nil {
			break
		}

		ws.Propagate(msg.Type, msg.Data)
	}
}
