package main

import (
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket is dumb - it echoes whatever it receives in both directions

type WebSocket struct {
	MessageBroadcaster
	conn    *websocket.Conn
	wsMutex sync.Mutex
}

func (ws *WebSocket) messageHandler(message *Message) error {
	ws.wsMutex.Lock()
	defer ws.wsMutex.Unlock()

	return ws.conn.WriteJSON(message)
}

func (ws *WebSocket) listen() {
	for {
		var msg Message
		err := ws.conn.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		ws.broadcast(&msg)
	}
}

// #TODO
/*
const (
	// Time allowed to write the file to the client.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the client.
	pongWait = 60 * time.Second

	// Send pings to client with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
)
*/

// Ensure you:
//
//	defer sender.Close()

/*

	http.Handle("/static/", http.StripPrefix("/static/", wsFunc))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		room, nick, err := getRoomAndNick(r.URL.Path)
		if err != nil {
			log.Println(err)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}

		log.Println("Client connected to room: '" + *room + "' with nick: '" + *nick + "'")
		client := &client{referee: referee, room: *room, nick: *nick, conn: conn, send: make(chan []byte, 256)}
		client.referee.register <- client
		go client.writeToConnectionPump()
		go client.readFromConnectionPump()
	})

	return &s, nil
}

func (cp *ClientPanel) wsRead() {
	for {
		var msg Msg
		err := cp.wsClient.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		default:
			log.Printf("Message not understood: %s\n", msg.Type)
		}
	}
}

func (cp *ClientPanel) wsWrite() {
	/*
		statusTicker := time.NewTicker(statusSendPeriod)
		defer func() {
			statusTicker.Stop()
		}()
*/
/*
	for {
		select {
		//		case <-done:
		//			return
		//		case <-statusTicker.C:
		//			fmt.Println("TICKER!")

		case status := <-cp.chSendServerStatus:
			msg := Msg{Type: "server-status", ServerStatus: status}
			err := cp.sendData(&msg)
			if err != nil {
				// #TODO: relax
				log.Fatal(err)
			}
		}
	}
}

/*
func (s *WebSocket) Close() error {
	msg := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	err := s.sendControlSignal(websocket.CloseMessage, msg, time.Second)
	if err != nil {
		// #TODO: log - though I don't know if we should still try close?
	}

	return s.conn.Close()
}

func (s *WebSocketLink) sendControlSignal(messageType int, data []byte, byDuration time.Duration) error {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()

	return s.conn.WriteControl(websocket.CloseMessage, data, time.Now().Add(byDuration))
}

func (s *WebSocketLink) sendData(data interface{}) error {
	s.wsMutex.Lock()
	defer s.wsMutex.Unlock()
	return s.conn.WriteJSON(data)
}
*/
