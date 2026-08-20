package websocket

import (
	"encoding/json"
	"fmt"
	"log"
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

type ReadHandler func(WebSocket) error

func reportError(chError chan<- error, err error) {
	if err == nil {
		return
	}

	select {
	case chError <- err:
	default:
		log.Printf("websocket error: %s", err)
	}
}

// Returns an HttpHandler that reads from a websocket connection
// fnOnConnOpen is an optional function that is called when the connection is opened
func NewWebSocketHandler(
	readFn ReadHandler,
	chError chan<- error,
	chSend <-chan message.Msg,
	fnOnConnOpen *func(),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := WsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			reportError(chError, err)
			return
		}
		defer conn.Close()

		ws := WebSocket{Conn: conn}
		done := make(chan struct{})
		defer close(done)

		go func() {
			for {
				select {
				case sendData, ok := <-chSend:
					if !ok {
						return
					}

					jsonData, err := json.Marshal(&sendData)
					if err != nil {
						reportError(chError, fmt.Errorf("marshal error: %s", err))
						return
					}

					if err := ws.Conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
						reportError(chError, err)
						_ = ws.Conn.Close()
						return
					}
				case <-done:
					return
				case <-r.Context().Done():
					return
				}
			}
		}()

		if fnOnConnOpen != nil {
			(*fnOnConnOpen)()
		}

		for {
			if err := readFn(ws); err != nil {
				reportError(chError, err)
				return
			}
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
	readerFn := func(ws WebSocket) error {
		messageType, msgRaw, err := ws.Conn.ReadMessage()
		if err != nil {
			return err
		}

		if messageType != websocket.BinaryMessage {
			return fmt.Errorf("messageType is not Binary")
		}

		// Unmarshal the header - if this fails we can't proceed
		var msgHeader message.MsgHeader
		err = json.Unmarshal(msgRaw, &msgHeader)
		if err != nil {
			return fmt.Errorf("header unmarshal: %s", err)
		}

		msgHandlerFn(msgHeader.Type, msgRaw)
		return nil
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
