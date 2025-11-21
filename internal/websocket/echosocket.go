package websocket

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
)

type WSTic80Data struct {
	S    string `json:"s"`  // "tic80"
	ID   string `json:"id"` // room/user
	Data string `json:"data"`
}

// EchoSocket attaches to a WebSocket server and echoes messages received to a channel
// TODO: clean up and error handling, etc
func Tic80SocketListener(socketURL url.URL, chReceived chan<- []byte) error {
	fmt.Printf("-> Connecting to %s", socketURL.String())

	conn, _, err := websocket.DefaultDialer.Dial(socketURL.String(), nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	//conn.SetReadLimit(maxMessageSize)
	//conn.SetReadDeadline(time.Now().Add(pongWait))
	//conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				//log.Printf("error: %v", err)
			}
			break
		}

		// TODO: Probably don't want to spam errors out like this!

		var wsData WSTic80Data
		err = json.Unmarshal(message, &wsData)
		if err != nil {
			continue
		}

		if wsData.S != "tic80" {
			fmt.Printf("Unknown WS data source: %s\n", wsData.S)
			continue
		}

		chReceived <- []byte(wsData.Data)
	}

	return nil
}
