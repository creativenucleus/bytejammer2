package controlpanel

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"

	"github.com/creativenucleus/bytejammer2/internal/basecontrolpanel"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
	"github.com/tyler-sommer/stick"
)

//go:embed page-templates/server-panel-index.html
var serverPanelIndexHtml []byte

type ServerPanel struct {
	basecontrolpanel.BaseControlPanel
}

func NewServerPanel(port uint) *ServerPanel {
	sp := ServerPanel{
		BaseControlPanel: *basecontrolpanel.NewControlPanel(port, fmt.Sprintf("Go to http://localhost:%d/", port)),
	}

	chError := make(chan error)
	chSend := make(chan message.Msg)

	router := sp.Router()

	router.HandleFunc("/", sp.serverPanelIndex)
	router.HandleFunc("/ws-server", websocket.NewWebSocketMsgHandler(func(msgType message.MsgType, msgData []byte) {
		switch msgType {
		default:
			fmt.Printf("Message not understood: %s\n", msgType)
		}
	}, chError, chSend))

	return &sp
}

func (sp *ServerPanel) serverPanelIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(serverPanelIndexHtml), w, nil)
	if err != nil {
		log.Println("write:", err)
	}
}
