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

//go:embed page-templates/obs-overlay.html
var ObsOverlayIndexHtml []byte

type ObsOverlay struct {
	ControlPanel
	chSend chan string
}

func NewObsOverlay(
	port uint,
	//	chMakeSnapshot chan<- message.MsgDataMakeSnapshot,
	//	chNewPlayer chan<- bool,
) *ObsOverlay {
	panel := ObsOverlay{
		ControlPanel: *NewControlPanel(port, fmt.Sprintf("Go to http://localhost:%d/", port)),
	}

	chError := make(chan error)
	panel.chSend = make(chan string)

	panel.router.HandleFunc("/", panel.webIndex)
	panel.router.HandleFunc("/ws-obs-overlay",
		websocket.NewWebSocketMsgHandler(func(msgType message.MsgType, msgData []byte) {
			switch msgType {

			default:
				fmt.Printf("Message not understood: %s\n", msgType)
			}
		}, chError, panel.chSend),
	)

	return &panel
}

func (cp *ObsOverlay) webIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(ObsOverlayIndexHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}

func (o *ObsOverlay) SetDetail(playerName string, effectName string) error {
	fmt.Printf("Setting OBS overlay details: Player Name: %s, Effect Name: %s\n", playerName, effectName)
	// Placeholder for the actual implementation of setting details in the OBS overlay.
	// This function should update the overlay with the provided player name and effect name.
	// TODO: sanitise
	o.chSend <- fmt.Sprintf(
		`<div class="authorName">Author: %s</div><div class="effectName">Effect: %s</div>`,
		playerName, effectName,
	)

	return nil
}
