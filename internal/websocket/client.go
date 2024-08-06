package websocket

import (
	"fmt"
	"log"
	"net/url"

	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/gorilla/websocket"
)

// Implements message.MsgReceiver
type WebSocket struct {
	message.MsgSender
	conn *websocket.Conn
	//	wsMutex sync.Mutex
}

// Try to join host:port/path with a websocket connection
func NewWebSocketClient(host string, port int, path string) (*WebSocket, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   path,
	}
	log.Printf("-> Connecting to %s", u.String())

	ws := &WebSocket{}
	var err error
	ws.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	return ws, nil
}

// WebSocket is dumb - it echoes whatever it receives in both directions

func (ws *WebSocket) MsgHandler(msg message.Msg) error {
	//	ws.wsMutex.Lock()
	//	defer ws.wsMutex.Unlock()
	switch msg.Type {
	case message.MsgTypeTicState:
		ticState, ok := msg.Data.(*tic.State)
		if !ok {
			return nil
		}

		data, err := ticState.MakeDataToImport()
		if err != nil {
			fmt.Println(err)
			return err
		}

		return ws.conn.WriteMessage(websocket.TextMessage, data)
	}

	return nil
}

/*
func (ws *WebSocket) listen() {
	for {
		var msg message.Msg
		err := ws.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		ws.Send(msg)
	}
}
*/
