package websocket

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/gorilla/websocket"
)

type WebSocketServer struct {
	message.MsgPropagator
	conn *websocket.Conn
	//	wsMutex      sync.Mutex
	//	chLog        chan string
	//	statusTicker *time.Ticker
}

// endpoint: /myendpoint
func NewWebSocketServer(port int, endpoint string) (*WebSocketServer, error) {
	wss := WebSocketServer{}

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	http.HandleFunc(endpoint, wss.socketHandler())
	// #TODO: unyuck this
	go func() {
		err := webServer.ListenAndServe()
		if err != nil {
			fmt.Println("ERROR")
			//			return err
		}
	}()

	return &wss, nil
}

func (wss *WebSocketServer) socketHandler() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		wss.conn, err = WsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}

		defer func() {
			wss.conn.Close()
			wss.conn = nil
		}()

		go wss.wsOperatorRead()
		//		go wss.wsOperatorWrite()

		// #TODO: handle exit
		for {
			time.Sleep(time.Second)
		}
	}
}

func (wss *WebSocketServer) wsOperatorRead() {
	propagateIncomingMessages(wss.conn, func(msgType message.MsgType, msgData []byte) {
		wss.Propagate(msgType, msgData)
	})
}

/*
func (wss *WebSocketServer) wsOperatorWrite() {
	wss.statusTicker = time.NewTicker(statusSendPeriod)
	defer func() {
		wss.statusTicker.Stop()
	}()

	for {
		select {
		//		case <-done:
		//			return
		case <-wss.statusTicker.C:
			wss.sendServerStatus(false)

		case logMsg := <-wss.chLog:
			wss.sendLog(logMsg)
		}
	}
}
*/
