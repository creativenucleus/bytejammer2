package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/gorilla/websocket"
)

// Implements message.MsgReceiver
type WebSocketRawData struct {
	message.MsgPropagator
	conn *websocket.Conn
	//	wsMutex sync.Mutex
}

// Try to join host:port/path with a websocket connection
func NewWebSocketRawDataClient(host string, port int, path string) (*WebSocketRawData, error) {
	u := url.URL{
		Scheme: "ws",
		Host:   fmt.Sprintf("%s:%d", host, port),
		Path:   path,
	}
	log.Printf("-> Connecting to %s", u.String())

	ws := &WebSocketRawData{}
	var err error
	ws.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}

	//	go ws.listen()

	return ws, nil
}

// WebSocket is dumb - it echoes whatever it receives in both directions

func (ws *WebSocketRawData) MsgHandler(msgType message.MsgType, msgData []byte) error {
	//	ws.wsMutex.Lock()
	//	defer ws.wsMutex.Unlock()
	switch msgType {
	case message.MsgTypeTicState:
		var ticState tic.State
		err := json.Unmarshal(msgData, &ticState)
		if err != nil {
			return err
		}

		data, err := ticState.MakeDataToImport()
		if err != nil {
			return err
		}

		return ws.conn.WriteMessage(websocket.TextMessage, data)
	case message.MsgTypeTicSnapshot:
		fmt.Println("OH!")
	}

	return nil
}

/*
func (ws *WebSocketRawData) listen() {
	for {
		messageType, msg, err := ws.conn.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		fmt.Println(messageType)

		state := tic.State{}
		state.SetCode(data)
		data, err = json.Marshal(state)
		if err != nil {
			log.Println("error marshalling state")
			break
		}

		ws.Propagate(message.Msg{Type: message.MsgTypeTicState, Data: data})
	}
}
*/

/*
func (ws *WebSocketRawData) listen() {
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
