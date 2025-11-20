package studio

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/creativenucleus/bytejammer2/internal/basecontrolpanel"
	"github.com/creativenucleus/bytejammer2/internal/controlpanel"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
	"github.com/tyler-sommer/stick"
)

//go:embed page-templates/studio-index.html
var studioIndexHtml []byte

type ticSocketWatcher struct {
	listenToUrl string
	overlayPort uint
	playerName  string
}

type Studio struct {
	controlPanel      basecontrolpanel.BaseControlPanel
	ticSocketWatchers []ticSocketWatcher
}

func NewStudio(
	chUserExitRequest <-chan bool,
	port uint,
) *Studio {
	controlPanel := *basecontrolpanel.NewControlPanel(port, fmt.Sprintf("Go to http://localhost:%d/", port))

	studio := Studio{
		controlPanel: controlPanel,
	}

	chError := make(chan error)
	chSend := make(chan message.Msg)

	router := studio.controlPanel.Router()

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		env := stick.New(nil)

		err := env.Execute(string(studioIndexHtml), w, map[string]stick.Value{})
		if err != nil {
			log.Println("write:", err)
		}
	})

	router.HandleFunc("/action/launch-tic-with-overlay.json", func(w http.ResponseWriter, r *http.Request) {
		fmt.Printf("%+v\n", r.Body)

		var body message.MsgDataStartTicWithOverlay
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"could not decode request: %s"}`, err)
			return
		}

		// TODO: change!
		go func() {
			destFilePath := body.FileStub
			config := controlpanel.ObsOverlayServerConfig{
				ProxyDestFile:  destFilePath,
				PlayerName:     body.PlayerName,
				ObsOverlayPort: body.OverlayPort,
			}

			// TODO: faked!
			chWebsocketWatcher := make(<-chan []byte)

			err = controlpanel.ObsOverlayRun(chUserExitRequest, config, chWebsocketWatcher)
			if err != nil {
				fmt.Printf("Error running overlay: %s\n", err)
				return
			}
		}()

		// Fake for now
		studio.ticSocketWatchers = append(studio.ticSocketWatchers, ticSocketWatcher{
			listenToUrl: body.ListenToUrl,
			overlayPort: body.OverlayPort,
			playerName:  body.PlayerName,
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			struct {
				Success bool `json:"success"`
			}{
				Success: true,
			},
		)

		// Send Status via websocket
		chSend <- message.Msg{
			Type: message.MsgTypeStudioServerStatus,
			Data: map[string]any{
				"running_count": studio.ticSocketWatchers,
			},
		}
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

	return &studio
}

func (s *Studio) Launch() error {
	return s.controlPanel.Launch()
}
