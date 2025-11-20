package basecontrolpanel

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/gorilla/mux"
)

//go:embed serve-static/*
var webStaticAssets embed.FS

type BaseControlPanel struct {
	port   uint
	router *mux.Router
	//	wsClient *websocket.WebSocketRawData
	//	wsMutex  sync.Mutex
	startedMessage string
}

func NewControlPanel(port uint, startedMessage string) *BaseControlPanel {
	cp := BaseControlPanel{
		port:           port,
		router:         mux.NewRouter(),
		startedMessage: startedMessage,
	}
	return &cp
}

func (cp *BaseControlPanel) Launch() error {
	subFs, err := fs.Sub(webStaticAssets, "serve-static")
	if err != nil {
		return err
	}
	//	cp.router.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFs))))
	cp.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(subFs))))

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", cp.port),
		Handler:           cp.router,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
	}

	log.GlobalLog.Log("info", cp.startedMessage)

	//	http.HandleFunc(fmt.Sprintf("/%s", session), cp.webClientIndex)

	//	http.HandleFunc(fmt.Sprintf("/%s/api/identity.json", session), cp.webClientApiIdentityJSON)
	//	http.HandleFunc(fmt.Sprintf("/%s/api/join-server.json", session), cp.webClientApiJoinServerJSON)

	//	http.HandleFunc(fmt.Sprintf("/%s/ws-client", session), cp.wsWebClient())
	return webServer.ListenAndServe()
}

func (cp *BaseControlPanel) Router() *mux.Router {
	return cp.router
}

/*
func (cp *ControlPanel) webClientIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(ClientIndexHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.GlobalLog.Log("info", fmt.Sprintf("write: %s", err))
	}
}
*/
/*
func (cp *ControlPanel) wsWebClient() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var err error
		cp.wsClient, err = WsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer cp.wsClient.Close()

		go cp.wsRead()
		go cp.wsWrite()

		// #TODO: handle exit
		for {
		}
	}
}

func (cp *ControlPanel) wsRead() {
	for {
		var msg Msg
		err := cp.wsClient.ReadJSON(&msg)
		if err != nil {
			log.Println("read:", err)
			break
		}

		switch msg.Type {
		default:
			log.Printf("Message not understood: %s\n", msg.Type)
		}
	}
}

func (cp *ControlPanel) wsWrite() {
	/*
		statusTicker := time.NewTicker(statusSendPeriod)
		defer func() {
			statusTicker.Stop()
		}()
*/
/*
	for {
		select {
		//		case <-done:
		//			return
		//		case <-statusTicker.C:
		//			fmt.Println("TICKER!")

		case status := <-cp.chSendServerStatus:
			msg := Message{Type: "server-status", ServerStatus: status}
			err := cp.sendData(&msg)
			if err != nil {
				// #TODO: relax
				log.Fatal(err)
			}
		}
	}
}

func (cp *ControlPanel) sendData(data interface{}) error {
	cp.wsMutex.Lock()
	defer cp.wsMutex.Unlock()
	return cp.wsClient.WriteJSON(data)
}
*/
