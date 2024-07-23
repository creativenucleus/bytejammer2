package controlpanel

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"time"
)

//go:embed serve-static/*
var WebStaticAssets embed.FS

type ControlPanel struct {
}

func Start(port int) (*ControlPanel, error) {
	// Replace this with a random string...
	session := "session"

	webServer := &http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		ReadHeaderTimeout: 3 * time.Second,
	}

	subFs, err := fs.Sub(WebStaticAssets, "web-static")
	if err != nil {
		return nil, err
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(subFs))))

	fmt.Printf("In a web browser, go to http://localhost:%d/%s\n", port, session)

	cp := ControlPanel{
		//		chSendServerStatus: make(chan ClientServerStatus),
	}
	//	http.HandleFunc(fmt.Sprintf("/%s", session), cp.webClientIndex)
	//	http.HandleFunc(fmt.Sprintf("/%s/api/identity.json", session), cp.webClientApiIdentityJSON)
	//	http.HandleFunc(fmt.Sprintf("/%s/api/join-server.json", session), cp.webClientApiJoinServerJSON)
	//	http.HandleFunc(fmt.Sprintf("/%s/ws-client", session), cp.wsWebClient())
	if err := webServer.ListenAndServe(); err != nil {
		return nil, err
	}

	return &cp, nil
}
