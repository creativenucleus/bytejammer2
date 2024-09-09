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

//go:embed page-templates/server-panel-index.html
var ServerPanelIndexHtml []byte

type ServerPanel struct {
	ControlPanel
}

func NewServerPanel(
	port uint,
) *ServerPanel {
	sp := ServerPanel{
		ControlPanel: *NewControlPanel(port, fmt.Sprintf("Go to http://localhost:%d/", port)),
	}

	chError := make(chan error)

	sp.router.HandleFunc("/", sp.serverPanelIndex)
	sp.router.HandleFunc("/ws-server", websocket.NewWebSocketMsgHandler(func(msgType message.MsgType, msgData []byte) {
		switch msgType {
		default:
			fmt.Printf("Message not understood: %s\n", msgType)
		}
	}, chError))

	return &sp
}

func (sp *ServerPanel) serverPanelIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(ServerPanelIndexHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}
