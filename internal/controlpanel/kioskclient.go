package controlpanel

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
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

	chError := make(chan error)
	chSend := make(chan message.Msg)

	kc.router.HandleFunc("/", kc.webKioskIndex)
	kc.router.HandleFunc("/ws-kiosk",
		websocket.NewWebSocketMsgHandler(
			func(msgType message.MsgType, msgRaw []byte) {
				switch msgType {
				case message.MsgTypeKioskMakeSnapshot:
					body := struct {
						Data message.MsgDataMakeSnapshot `json:"data"`
					}{}
					err := json.Unmarshal(msgRaw, &body)
					if err != nil {
						fmt.Printf("Error unmarshalling data: %s\n", err)
						return
					}

					chMakeSnapshot <- body.Data

				case message.MsgTypeKioskNewPlayer:
					chNewPlayer <- true

				default:
					fmt.Printf("Message not understood: %s\n", msgType)
				}
			},
			chError,
			chSend,
		),
	)

	return &kc
}

func (cp *KioskClient) webKioskIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(KioskIndexHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}
