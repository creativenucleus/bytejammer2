package controlpanel

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
	gorillaWS "github.com/gorilla/websocket"
	"github.com/tyler-sommer/stick"
)

//go:embed page-templates/kiosk-index.html
var KioskIndexHtml []byte

type KioskClient struct {
	ControlPanel
}

func NewKioskClient(
	port uint,
	chMakeSnapshot chan<- message.MsgDataMakeSnapshot,
	chNewPlayer chan<- bool,
) *KioskClient {
	kc := KioskClient{
		ControlPanel: *NewControlPanel(port, fmt.Sprintf("Go to http://localhost:%d/", port)),
	}

	kc.router.HandleFunc("/", kc.webKioskIndex)
	kc.router.HandleFunc("/ws-kiosk", websocket.NewWebSocketHandler(func(ws websocket.WebSocket) {

		messageType, msgData, err := ws.Conn.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			return
		}

		if messageType != gorillaWS.BinaryMessage {
			fmt.Println("messageType is not Binary")
			return
		}

		var msgHeader message.MsgHeader
		err = json.Unmarshal(msgData, &msgHeader)
		if err != nil {
			fmt.Printf("Error unmarshalling header: %s\n", err)
			return
		}

		switch msgHeader.Type {
		case message.MsgTypeKioskMakeSnapshot:
			body := struct {
				Data message.MsgDataMakeSnapshot `json:"data"`
			}{}
			err := json.Unmarshal(msgData, &body)
			if err != nil {
				fmt.Printf("Error unmarshalling data: %s\n", err)
				return
			}

			chMakeSnapshot <- body.Data

		case message.MsgTypeKioskNewPlayer:
			chNewPlayer <- true

		default:
			fmt.Printf("Message not understood: %s\n", msgHeader.Type)
		}
	}))

	return &kc
}

func (cp *KioskClient) webKioskIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(KioskIndexHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}
