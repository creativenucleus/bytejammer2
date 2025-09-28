package controlpanel

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
	"github.com/tyler-sommer/stick"
)

//go:embed page-templates/obs-overlay-code.html
var obsOverlayCodeIndexHtml []byte

type ObsOverlayCode struct {
	ControlPanel
	chSend chan message.Msg
}

func NewObsOverlayCode(
	port uint,
	//	chMakeSnapshot chan<- message.MsgDataMakeSnapshot,
	//	chNewPlayer chan<- bool,
) *ObsOverlayCode {
	panel := ObsOverlayCode{
		ControlPanel: *NewControlPanel(port, fmt.Sprintf("Go to http://localhost:%d/", port)),
	}

	chError := make(chan error)
	panel.chSend = make(chan message.Msg)

	panel.router.HandleFunc("/", panel.webIndex)
	panel.router.HandleFunc("/ws-obs-overlay-code",
		websocket.NewWebSocketMsgHandler(
			func(msgType message.MsgType, msgRaw []byte) {
				switch msgType {

				default:
					fmt.Printf("Message not understood: %s\n", msgType)
				}
			},
			chError,
			panel.chSend,
		),
	)

	return &panel
}

func (cp *ObsOverlayCode) webIndex(w http.ResponseWriter, r *http.Request) {
	env := stick.New(nil)

	err := env.Execute(string(obsOverlayCodeIndexHtml), w, map[string]stick.Value{"session_key": "session"})
	if err != nil {
		log.Println("write:", err)
	}
}

func (o *ObsOverlayCode) SetCode(state tic.State, playerName string, isEditorUpdated bool) error {
	// TODO: sanitise code?
	splitCode := strings.Split(strings.ReplaceAll(string(state.Code), "\r\n", "\n"), "\n")
	if state.CursorY > 0 && state.CursorY <= len(splitCode) {
		line := splitCode[state.CursorY-1]
		lineLen := len(line)
		leftPos := state.CursorX - 1
		if state.CursorX > 0 && leftPos <= lineLen {
			oldLine := line

			line = oldLine[:leftPos]
			line = line + `<span class="cursor">`
			if leftPos < lineLen {
				line = line + oldLine[leftPos:leftPos+1]
			} else {
				line = line + " "
			}
			line = line + "</span>"
			if leftPos+1 < lineLen {
				line = line + oldLine[leftPos+1:]
			}
			splitCode[state.CursorY-1] = line
		}
	}
	rejoinedCode := strings.Join(splitCode, "\n")

	// TODO: sanitise
	playerNameHtml := ""
	if playerName != "" {
		playerNameHtml = fmt.Sprintf(`<div class="playerName">%s</div>`, playerName)
	}

	o.chSend <- message.Msg{
		Type: "obs-overlay-html",
		StringData: fmt.Sprintf(
			`%s<div id="codeContainer"><div class="code inactive-fade">%s</div></div>`,
			playerNameHtml,
			rejoinedCode,
		),
		Data: map[string]any{
			"isEditorUpdated": isEditorUpdated,
		},
	}

	return nil
}
