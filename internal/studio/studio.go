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
	"github.com/google/uuid"
	"github.com/gosimple/slug"
	"github.com/tyler-sommer/stick"
)

//go:embed page-templates/studio-index.html
var studioIndexHtml []byte

type ticRunner struct {
	id             uuid.UUID
	listenToURL    string
	playerName     string
	slug           string
	filePath       string
	overlayURLPath string
}

type Studio struct {
	server     *webserver.Webserver
	hostPart   string
	ticRunners []ticRunner
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
	hostPart := fmt.Sprintf("http://localhost:%d", port)

	logMessage := fmt.Sprintf("Starting Studio control panel on %s", hostPart)
	server, err := webserver.NewWebserver(port, logMessage)
	if err != nil {
		log.Fatalf("Could not create webserver: %s", err)
	}

	studio := Studio{}
	studio.server = server
	studio.hostPart = hostPart
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

	router.HandleFunc("/action/start-tic-runner.json", func(w http.ResponseWriter, r *http.Request) {
		statusCode, successMessage, err := studio.handleStartTicRunner(r)
		if err != nil {
			w.WriteHeader(statusCode)
			// TODO: Does this need to be escaped?
			fmt.Fprintf(w, `{"message": "%s"}`, err)
			return
		}

		studio.sendLog("success", successMessage)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}{
				Success: true,
				Message: successMessage,
			},
		)

		studio.sendServerStatus()
	}).Methods("POST")

	router.HandleFunc("/action/stop-tic-runner.json", func(w http.ResponseWriter, r *http.Request) {
		var body message.MsgDataStopTicRunner
		err := json.NewDecoder(r.Body).Decode(&body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, `{"error":"could not decode request: %s"}`, err)
			return
		}

		err = studio.stopTicRunner(body.ID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{"error":"could not stop TIC runner: %s"}`, err)
			return
		}

		responseMessage := "TIC runner: stopped"
		studio.sendLog("success", responseMessage)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(
			struct {
				Success bool   `json:"success"`
				Message string `json:"message"`
			}{
				Success: true,
				Message: responseMessage,
			},
		)

		studio.sendServerStatus()
	}).Methods("POST")

	onConnOpen := func() {
		studio.sendServerStatus()
	}

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
			&onConnOpen,
		),
	)

	// Chew through errors and send to the web panel
	// (TODO: This may be problematic if there are too many errors or the socket is broken?)
	go func() {
		for err := range studio.chError {
			studio.sendLog("error", err.Error())
		}
	}()

	return &studio
}

func (s *Studio) Run() error {
	return s.server.Run()
}

// Returns http status code and [successMessage or error]
func (s *Studio) handleStartTicRunner(r *http.Request) (int, string, error) {
	var body message.MsgDataStartTicRunner
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		return http.StatusBadRequest, "", fmt.Errorf("could not decode request: %s", err)
	}

	slug := slug.Make(fmt.Sprintf("tic-%s", body.PlayerName)) // TODO: make unique!
	if slug == "" {
		return http.StatusBadRequest, "", fmt.Errorf("could not create slug from player name: %s", body.PlayerName)
	}

	for _, existingRunner := range s.ticRunners {
		if existingRunner.slug == slug {
			return http.StatusBadRequest, "", fmt.Errorf("a TIC runner with slug '%s' already exists", slug)
		}
	}

	switch body.ObsOverlay {
	case "none":
		_, err = s.addTicRunner(slug, body.ListenToUrl, body.PlayerName)
		if err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("could not add TIC runner: %s", err)
		}

	case "code":
		_, err = s.addTicRunnerWithOverlay(slug, body.ListenToUrl, body.PlayerName)
		if err != nil {
			return http.StatusInternalServerError, "", fmt.Errorf("could not add TIC runner with overlay: %s", err)
		}

	default:
		return http.StatusBadRequest, "", fmt.Errorf("unknown obsOverlay option: %s", body.ObsOverlay)
	}

	return http.StatusOK, fmt.Sprintf("TIC runner started for player: '%s'", body.PlayerName), nil
}

// addTicRunner adds a new socket watcher, outputting to a TIC
// listenToURL is the URL to listen to TIC data from
// playerName is the name of the player to associate with this watcher
// The playerName will be used to create a file for the TIC to watch
// Returns a ticRunner
func (s *Studio) addTicRunner(slug string, listenToURL string, playerName string) (*ticRunner, error) {
	fileDir := filepath.Join(config.CONFIG.WorkDir, "bytejam")
	err := files.EnsurePathExists(fileDir, 0755)
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(fileDir, slug+".ticcode")

	// TODO: Ensure we aren't duplicating slugs

	chReceived := make(chan []byte)

	// TODO: this is all quite dicey

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
		err = TicRunner(s.chUserExitRequest, chReceived, filePath)
		if err != nil {
			fmt.Printf("Error running overlay: %s\n", err)
		}
	}()

	ticRunner := ticRunner{
		id:          uuid.New(),
		listenToURL: listenToURL,
		playerName:  playerName,
		slug:        slug,
		filePath:    filePath,
	}

	// Fake for now
	s.ticRunners = append(s.ticRunners, ticRunner)

	return &ticRunner, nil
}

// addTicRunnerWithOverlay adds a new socket watcher, outputting to overlay and a TIC
// listenToURL is the URL to listen to TIC data from
// playerName is the name of the player to associate with this watcher
// The playerName will be used to launch a URL, and a file for the TIC to watch
// Returns a ticRunner
func (s *Studio) addTicRunnerWithOverlay(slug string, listenToURL string, playerName string) (*ticRunner, error) {
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
		err = TicOverlayRunner(s.chUserExitRequest, chReceived, codeOverlayPanel, filePath)
		if err != nil {
			fmt.Printf("Error running overlay: %s\n", err)
		}
	}()

	ticRunner := ticRunner{
		listenToURL:    listenToURL,
		playerName:     playerName,
		slug:           slug,
		filePath:       filePath,
		overlayURLPath: overlayURLPath,
	}

	// Fake for now
	s.ticRunners = append(s.ticRunners, ticRunner)

	return &ticRunner, nil
}

func (s *Studio) stopTicRunner(id uuid.UUID) error {
	newRunners := []ticRunner{}
	for _, runner := range s.ticRunners {
		if runner.id == id {
			// TODO: actually stop the runner!
		} else {
			// It doesn't match our ID, so we keep this one
			newRunners = append(newRunners, runner)
		}
	}
	s.ticRunners = newRunners
	return nil
}

// sendServerStatus sends the current server status to all connected websocket clients
func (s *Studio) sendServerStatus() {
	type statusOverlay struct {
		ID             uuid.UUID `json:"id"`
		PlayerName     string    `json:"playerName"`
		ListenToURL    string    `json:"listenToURL"`
		OverlayURL     string    `json:"overlayURL"`
		OverlayURLPath string    `json:"overlayURLPath"`
		FilePath       string    `json:"filePath"`
	}

	var overlays []statusOverlay
	for _, ticRunner := range s.ticRunners {
		// TODO: raise if there's an error
		fullFilePath, _ := filepath.Abs(ticRunner.filePath)

		overlayURL := ""
		if ticRunner.overlayURLPath != "" {
			overlayURL = s.hostPart + ticRunner.overlayURLPath
		}

		overlays = append(overlays, statusOverlay{
			ID:             ticRunner.id,
			PlayerName:     ticRunner.playerName,
			ListenToURL:    ticRunner.listenToURL,
			OverlayURL:     overlayURL,
			OverlayURLPath: ticRunner.overlayURLPath,
			FilePath:       fullFilePath,
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

// sendServerStatus sends the current server status to all connected websocket clients
func (s *Studio) sendLog(level string, messageText string) {
	// Send Log via websocket
	s.chWSSend <- message.Msg{
		Type: message.MsgTypeLog,
		Data: map[string]any{
			"level":   level,
			"message": messageText,
		},
	}
}
