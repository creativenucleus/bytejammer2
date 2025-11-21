package obs

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/creativenucleus/bytejammer2/internal/message"
	"github.com/creativenucleus/bytejammer2/internal/tic"
	"github.com/creativenucleus/bytejammer2/internal/websocket"
	"github.com/gorilla/mux"
	"github.com/tyler-sommer/stick"
)

//go:embed page-templates/obs-overlay-code.html
var overlayCodePanelIndexHtml []byte

type CodeOverlayPanel struct {
	pathOverlay string
	chSendCode  chan message.Msg
	// playerName is infused into the overlay HTML (this could be changed)
	playerName string
}

func NewCodeOverlayPanel(
	router *mux.Router,
	pathOverlay string,
	pathWSOverlay string,
	playerName string,
	chError chan error,
) (*CodeOverlayPanel, error) {
	panel := CodeOverlayPanel{}

	panel.chSendCode = make(chan message.Msg)
	panel.pathOverlay = pathOverlay
	panel.playerName = playerName

	router.HandleFunc(panel.pathOverlay, func(w http.ResponseWriter, r *http.Request) {
		env := stick.New(nil)

		err := env.Execute(
			string(overlayCodePanelIndexHtml),
			w,
			map[string]stick.Value{
				"ws_path": pathWSOverlay,
			},
		)
		if err != nil {
			log.Println("write:", err)
		}
	})

	router.HandleFunc(pathWSOverlay,
		websocket.NewWebSocketMsgHandler(
			func(msgType message.MsgType, msgRaw []byte) {
				switch msgType {
				// No incoming messages expected
				default:
					fmt.Printf("Message not understood: %s\n", msgType)
				}
			},
			chError,
			panel.chSendCode,
			nil,
		),
	)

	return &panel, nil
}

func (p *CodeOverlayPanel) SetCode(state tic.State, isEditorUpdated bool) error {
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
	if p.playerName != "" {
		playerNameHtml = fmt.Sprintf(`<div class="playerName">%s</div>`, p.playerName)
	}

	p.chSendCode <- message.Msg{
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
