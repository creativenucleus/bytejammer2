package controlpanel

import (
	_ "embed"
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

func NewKioskClient(port uint, chMakeSnapshot chan<- bool) *KioskClient {
	kc := KioskClient{
		ControlPanel: *NewControlPanel(port, fmt.Sprintf("Go to http://localhost:%d/", port)),
	}

	kc.router.HandleFunc("/", kc.webKioskIndex)
	kc.router.HandleFunc("/ws-kiosk", websocket.NewWebSocketHandler(func(ws websocket.WebSocket) {
		var msg message.Msg
		err := ws.Conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("read:", err)
			return
		}

		switch msg.Type {
		case message.MsgTypeKioskMakeSnapshot:
			fmt.Println("Make snapshot")
			chMakeSnapshot <- true

		default:
			fmt.Printf("Message not understood: %s\n", msg.Type)
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
