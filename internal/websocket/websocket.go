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
// fnOnConnOpen is an optional function that is called when the connection is opened
func NewWebSocketHandler(
	readFn ReadHandler,
	chError chan<- error,
	chSend <-chan message.Msg,
	fnOnConnOpen *func(),
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
					jsonData, err := json.Marshal(&sendData)
					if err != nil {
						chError <- fmt.Errorf("marshal error: %s", err)
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

		if fnOnConnOpen != nil {
			(*fnOnConnOpen)()
		}

		// Receive
		for {
			readFn(ws, chError)
		}
	}
}

type MsgHandlerFn func(msgType message.MsgType, msgRaw []byte)

// Returns an HttpHandler that reads messages in our format from a websocket connection
func NewWebSocketMsgHandler(
	msgHandlerFn MsgHandlerFn,
	chError chan<- error,
	chSend <-chan message.Msg,
	fnOnConnOpen *func(),
) func(w http.ResponseWriter, r *http.Request) {
	readerFn := func(ws WebSocket, chError chan<- error) {
		messageType, msgRaw, err := ws.Conn.ReadMessage()
		if err != nil {
			chError <- err
			return
		}

		if messageType != websocket.BinaryMessage {
			chError <- fmt.Errorf("messageType is not Binary")
			return
		}

		// Unmarshal the header - if this fails we can't proceed
		var msgHeader message.MsgHeader
		err = json.Unmarshal(msgRaw, &msgHeader)
		if err != nil {
			chError <- fmt.Errorf("header unmarshal: %s", err)
			return
		}

		msgHandlerFn(msgHeader.Type, msgRaw)
	}

	return NewWebSocketHandler(readerFn, chError, chSend, fnOnConnOpen)
}

// #TODO: Make less brittle
// Listens to the incoming messages and propagates the,
func propagateIncomingMessages(conn *websocket.Conn, propagate MsgHandlerFn) error {
	for {
		socketMsgType, socketMsgData, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				//				log.Println("Connection unexpectedly closed")
				return err
			}

			//			log.Println("unhandled socket read error:", err)
			return err
		}

		if socketMsgType != websocket.TextMessage {
			//			log.Println("messageType is not Text")
			continue
		}

		var msg message.Msg
		err = json.Unmarshal(socketMsgData, &msg)
		if err != nil {
			break
		}

		propagate(msg.Type, socketMsgData)
	}

	return nil
}
