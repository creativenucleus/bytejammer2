package controlpanel

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/creativenucleus/bytejammer2/internal/basecontrolpanel"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
	"github.com/tyler-sommer/stick"
)

//go:embed page-templates/studio-index.html
var studioIndexHtml []byte

type StudioPanel struct {
	basecontrolpanel.BaseControlPanel
}

func NewStudioPanel(
	chUserExitRequest <-chan bool,
	port uint,
) *StudioPanel {
	panel := StudioPanel{
		BaseControlPanel: *basecontrolpanel.NewControlPanel(port, fmt.Sprintf("Go to http://localhost:%d/", port)),
	}

	chError := make(chan error)
	chSend := make(chan message.Msg)

	router := panel.Router()

	router.HandleFunc("/", panel.webIndex)

	router.HandleFunc("/action/launch-tic-with-overlay.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%+v\n", r.Body)

		body := struct {
			Data message.MsgDataStartTicWithOverlay `json:"data"`
		}{}
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"could not decode request: %s"}`, err)
			return
		}

		fmt.Println("Launching TIC with overlay with params:")
		fmt.Printf("%+v\n", body.Data)

		// TODO: change!
		destFilePath := body.Data.FileStub
		config := ObsOverlayServerConfig{
			ProxySourceFile: body.Data.FileStub,
			ProxyDestFile:   destFilePath,
			PlayerName:      body.Data.PlayerName,
			ObsOverlayPort:  body.Data.OverlayPort,
		}

		err = ObsOverlayRun(chUserExitRequest, config)
		if err != nil {
			fmt.Printf("Error running overlay: %s\n", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		)
	}).Methods("POST")

	router.HandleFunc("/ws",
		websocket.NewWebSocketMsgHandler(
			func(msgType message.MsgType, msgRaw []byte) {
				switch msgType {
				default:
					fmt.Printf("Message not understood: %s\n", msgType)
				}
			},
			chError,
			chSend,
		),
	)

	return &panel
}

func (cp *StudioPanel) webIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(studioIndexHtml), w, map[string]stick.Value{})
	if err != nil {
		log.Println("write:", err)
	}
}
