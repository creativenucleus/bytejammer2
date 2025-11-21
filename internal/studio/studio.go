package studio

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/creativenucleus/bytejammer2/config"
	"github.com/creativenucleus/bytejammer2/internal/controlpanel/obs"
	"github.com/creativenucleus/bytejammer2/internal/files"
	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/webserver"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
	"github.com/creativenucleus/bytejammer2/internal/webstatic"
	"github.com/gosimple/slug"
	"github.com/tyler-sommer/stick"
)

//go:embed page-templates/studio-index.html
var studioIndexHtml []byte

type ticSocketWatcher struct {
	listenToURL string
	playerName  string
	slug        string
	filePath    string
	overlayURL  string
}

type Studio struct {
	server            *webserver.Webserver
	ticSocketWatchers []ticSocketWatcher
	// Channel to listen for user exit requests
	chUserExitRequest <-chan bool
	// Send errors to the on this channel (we can log them, or send them to the web panel)
	chError chan error
	// Send messages to the web panel with this channel
	chWSSend chan message.Msg
}

func NewStudio(
	chUserExitRequest <-chan bool,
	port uint,
) *Studio {
	logMessage := fmt.Sprintf("Starting Studio control panel on http://localhost:%d", port)
	server, err := webserver.NewWebserver(port, logMessage)
	if err != nil {
		log.Fatalf("Could not create webserver: %s", err)
	}

	studio := Studio{}
	studio.server = server
	studio.chUserExitRequest = chUserExitRequest
	studio.chError = make(chan error)
	studio.chWSSend = make(chan message.Msg)

	router := server.Router()

	err = server.StaticRoute(webstatic.FS(), webstatic.FSEmbedPath(), "/static/")
	if err != nil {
		log.Fatalf("Could not setup static route: %s", err)
	}

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		env := stick.New(nil)

		err := env.Execute(string(studioIndexHtml), w, map[string]stick.Value{})
		if err != nil {
			log.Println("write:", err)
		}
	})

	router.HandleFunc("/action/launch-tic-with-overlay.json", func(w http.ResponseWriter, r *http.Request) {
		var body message.MsgDataStartTicWithOverlay
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"could not decode request: %s"}`, err)
			return
		}

		_, err = studio.addTicSocketWatcher(body.ListenToUrl, body.PlayerName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"could not add TIC socket watcher: %s"}`, err)
			return
		}

		logMessage := fmt.Sprintf("Started TIC socket watcher for player '%s'", body.PlayerName)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}{
				Success: true,
				Message: logMessage,
			},
		)

		studio.sendServerStatus()
	}).Methods("POST")

	router.HandleFunc("/ws/studio",
		websocket.NewWebSocketMsgHandler(
			func(msgType message.MsgType, msgRaw []byte) {
				switch msgType {
				default:
					fmt.Printf("Message not understood: %s\n", msgType)
				}
			},
			studio.chError,
			studio.chWSSend,
		),
	)

	return &studio
}

func (s *Studio) Run() error {
	return s.server.Run()
}

// addTicSocketWatcher adds a new TIC socket watcher
// listenToURL is the URL to listen to TIC data from
// playerName is the name of the player to associate with this watcher
// The playerName will be used to launch a URL, and a file for the TIC to watch
// Returns the file slug
func (s *Studio) addTicSocketWatcher(listenToURL string, playerName string) (*ticSocketWatcher, error) {
	slug := slug.Make(fmt.Sprintf("tic-overlay-%s", playerName)) // TODO: make unique!
	if slug == "" {
		return nil, fmt.Errorf("could not create slug from player name: %s", playerName)
	}

	fileDir := filepath.Join(config.CONFIG.WorkDir, "bytejam")
	err := files.EnsurePathExists(fileDir, 0755)
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(fileDir, slug+".ticcode")

	// TODO: Ensure we aren't duplicating slugs

	chReceived := make(chan []byte)

	// TODO: this is all quite dicey

	overlayURLPath := fmt.Sprintf("/obs/%s/overlay", slug)
	overlayURLPathWS := fmt.Sprintf("/obs/%s/ws-overlay", slug)

	codeOverlayPanel, err := obs.NewCodeOverlayPanel(s.server.Router(), overlayURLPath, overlayURLPathWS, playerName, s.chError)
	if err != nil {
		s.chError <- fmt.Errorf("could not create OBS overlay code panel: %s", err)
		return nil, err
	}

	wsURL, err := url.Parse(listenToURL)
	if err != nil {
		return nil, err
	}

	go func() {
		err = websocket.Tic80SocketListener(*wsURL, chReceived)
		if err != nil {
			fmt.Printf("Error running socket echo: %s\n", err)
		}
	}()

	go func() {
		err = OverlayRunner(s.chUserExitRequest, chReceived, codeOverlayPanel, filePath)
		if err != nil {
			fmt.Printf("Error running overlay: %s\n", err)
		}
	}()

	socketWatcher := ticSocketWatcher{
		listenToURL: listenToURL,
		playerName:  playerName,
		slug:        slug,
		filePath:    filePath,
		overlayURL:  overlayURLPath,
	}

	// Fake for now
	s.ticSocketWatchers = append(s.ticSocketWatchers, socketWatcher)

	return &socketWatcher, nil
}

// sendServerStatus sends the current server status to all connected websocket clients
func (s *Studio) sendServerStatus() {
	type statusOverlay struct {
		PlayerName  string `json:"playerName"`
		ListenToURL string `json:"listenToURL"`
		OverlayURL  string `json:"overlayURL"`
		FilePath    string `json:"filePath"`
	}

	var overlays []statusOverlay
	for _, watcher := range s.ticSocketWatchers {
		overlays = append(overlays, statusOverlay{
			PlayerName:  watcher.playerName,
			ListenToURL: watcher.listenToURL,
			OverlayURL:  watcher.overlayURL,
			FilePath:    watcher.filePath,
		})
	}

	// Send Status via websocket
	s.chWSSend <- message.Msg{
		Type: message.MsgTypeStudioServerStatus,
		Data: map[string]any{
			"overlays": overlays,
		},
	}
}
