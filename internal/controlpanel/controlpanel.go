package controlpanel

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/websocket"
	"github.com/tyler-sommer/stick"
)

//go:embed serve-static/*
var WebStaticAssets embed.FS

//go:embed page-templates/index.html
var ClientIndexHtml []byte

type ControlPanel struct {
	wsClient *websocket.WebSocket
	wsMutex  sync.Mutex
}

func Start(port uint) (*ControlPanel, error) {
	// Replace this with a random string...
	session := "session"

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	subFs, err := fs.Sub(WebStaticAssets, "serve-static")
	if err != nil {
		return nil, err
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFs))))

	fmt.Printf("In a web browser, go to http://localhost:%d/%s\n", port, session)

	cp := ControlPanel{
		//		chSendServerStatus: make(chan ClientServerStatus),
	}

	http.HandleFunc(fmt.Sprintf("/%s", session), cp.webClientIndex)

	//	http.HandleFunc(fmt.Sprintf("/%s/api/identity.json", session), cp.webClientApiIdentityJSON)
	//	http.HandleFunc(fmt.Sprintf("/%s/api/join-server.json", session), cp.webClientApiJoinServerJSON)

	//	http.HandleFunc(fmt.Sprintf("/%s/ws-client", session), cp.wsWebClient())
	if err := webServer.ListenAndServe(); err != nil {
		return nil, err
	}

	return &cp, nil
}

func (cp *ControlPanel) webClientIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(ClientIndexHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}

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
