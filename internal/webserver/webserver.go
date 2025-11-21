package webserver

import (
	"fmt"
	"io/fs"
	"net/http"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/gorilla/mux"
)

type Webserver struct {
	server http.Server
	router mux.Router
	// Emit this message to the global log when we start the server
	logMessageOnStart string
}

// Configures a web server
func NewWebserver(port uint, logMessageOnStart string) (*Webserver, error) {
	s := Webserver{
		router:            *mux.NewRouter(),
		logMessageOnStart: logMessageOnStart,
	}

	s.server = http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           &s.router,
		ReadHeaderTimeout: 3 * time.Second,
		WriteTimeout:      15 * time.Second,
		ReadTimeout:       15 * time.Second,
	}

	return &s, nil
}

// StaticRoute sets up the static file serving route
// assetFS is the embedded filesystem containing the static assets
// assetRootDir is the root directory within the embedded filesystem to remove from the path (e.g. "serve-static")
// serveDir is the directory to serve from on the web server (e.g. "/static/")
func (s *Webserver) StaticRoute(assetFS fs.FS, assetRootDir string, serveDir string) error {
	subFs, err := fs.Sub(assetFS, assetRootDir)
	if err != nil {
		return err
	}
	s.router.PathPrefix(serveDir).Handler(http.StripPrefix(serveDir, http.FileServer(http.FS(subFs))))

	return nil
}

func (s *Webserver) Router() *mux.Router {
	return &s.router
}

// Run runs the web server
// (this is a blocking call - note the returns from ListenAndServe)
func (s *Webserver) Run() error {
	log.GlobalLog.Log("info", s.logMessageOnStart)

	return s.server.ListenAndServe()
}
