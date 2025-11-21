package basecontrolpanel

import (
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/webstatic"
	"github.com/gorilla/mux"
)

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
	subFs, err := fs.Sub(webstatic.FS(), webstatic.FSEmbedPath())
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

	return webServer.ListenAndServe()
}

func (cp *BaseControlPanel) Router() *mux.Router {
	return cp.router
}
