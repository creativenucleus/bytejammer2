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

//go:embed page-templates/obs-overlay-kiosk.html
var obsOverlayKioskIndexHtml []byte

type ObsOverlayKiosk struct {
	basecontrolpanel.BaseControlPanel
	chSend chan message.Msg
}

func NewObsOverlayKiosk(
	port uint,
	//	chMakeSnapshot chan<- message.MsgDataMakeSnapshot,
	//	chNewPlayer chan<- bool,
) *ObsOverlayKiosk {
	panel := ObsOverlayKiosk{
		BaseControlPanel: *basecontrolpanel.NewControlPanel(port, fmt.Sprintf("Go to http://localhost:%d/", port)),
	}

	chError := make(chan error)
	panel.chSend = make(chan message.Msg)

	router := panel.Router()

	router.HandleFunc("/", panel.webIndex)
	router.HandleFunc("/ws-obs-overlay-kiosk",
		websocket.NewWebSocketMsgHandler(func(msgType message.MsgType, msgData []byte) {
			switch msgType {

			default:
				fmt.Printf("Message not understood: %s\n", msgType)
			}
		}, chError, panel.chSend),
	)

	return &panel
}

func (cp *ObsOverlayKiosk) webIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(obsOverlayKioskIndexHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}

func (o *ObsOverlayKiosk) SetDetail(playerName string, effectName string) error {
	fmt.Printf("Setting OBS overlay details: Player Name: %s, Effect Name: %s\n", playerName, effectName)
	// Placeholder for the actual implementation of setting details in the OBS overlay.
	// This function should update the overlay with the provided player name and effect name.
	// TODO: sanitise
	o.chSend <- message.Msg{
		Type: "obs-overlay-html",
		StringData: fmt.Sprintf(
			`<div class="authorName">Author: %s</div><div class="effectName">Effect: %s</div>`,
			playerName, effectName,
		),
	}

	return nil
}
