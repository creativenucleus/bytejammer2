package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/creativenucleus/bytejammer2/internal/message"
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

type ReadHandler func(WebSocket, chan<- error)

// Returns an HttpHandler that reads from a websocket connection
func NewWebSocketHandler(
	readFn ReadHandler,
	chError chan<- error,
	chSend <-chan string,
) func(w http.ResponseWriter, r *http.Request) {
	ws := WebSocket{}

	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		ws.Conn, err = WsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			chError <- err
		}
		defer ws.Conn.Close()

		// #TODO: handle exit

		// Send
		go func() {
			for {
				select {
				case sendData := <-chSend:
					messageObject := message.Msg{
						Type:       "obs-overlay-html",
						StringData: sendData,
					}

					jsonData, err := json.Marshal(&messageObject)
					if err != nil {
						chError <- fmt.Errorf("unmarshal error: %s", err)
						return
					}

					err = ws.Conn.WriteMessage(websocket.TextMessage, jsonData)
					if err != nil {
						chError <- err
						return
					}
				default:
					continue
				}
			}
		}()

		// Receive
		for {
			readFn(ws, chError)
		}
	}
}

type MsgHandlerFn func(msgType message.MsgType, msgData []byte)

// Returns an HttpHandler that reads messages in our format from a websocket connection
func NewWebSocketMsgHandler(
	msgHandlerFn MsgHandlerFn,
	chError chan<- error,
	chSend <-chan string,
) func(w http.ResponseWriter, r *http.Request) {
	readerFn := func(ws WebSocket, chError chan<- error) {
		messageType, msgData, err := ws.Conn.ReadMessage()
		if err != nil {
			chError <- err
			return
		}

		if messageType != websocket.BinaryMessage {
			chError <- fmt.Errorf("messageType is not Binary")
			return
		}

		var msgHeader message.MsgHeader
		err = json.Unmarshal(msgData, &msgHeader)
		if err != nil {
			chError <- fmt.Errorf("header unmarshal: %s", err)
			return
		}

		msgHandlerFn(msgHeader.Type, msgData)
	}

	return NewWebSocketHandler(readerFn, chError, chSend)
}

// #TODO: Make less brittle
func propagateIncomingMessages(conn *websocket.Conn, propagate MsgHandlerFn) error {
	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				//				log.Println("Connection unexpectedly closed")
				return err
			}

			//			log.Println("unhandled socket read error:", err)
			return err
		}

		if messageType != websocket.TextMessage {
			//			log.Println("messageType is not Text")
			continue
		}

		var msg message.Msg
		err = json.Unmarshal(data, &msg)
		if err != nil {
			break
		}

		propagate(msg.Type, msg.Data)
	}

	return nil
}
