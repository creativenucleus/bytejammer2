package bytejam_obs

// Watches a file (potentially deposited by Ticws) and splits into an OBS overlay, and a file for TIC to watch

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/controlpanel"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/tic"
)

// should either have a client or a host set
type ServerConfig struct {
	ProxySourceFile string
	ProxyDestFile   string
	ObsOverlayPort  uint
}

func Run(chUserExitRequest <-chan bool, conf ServerConfig) error {
	log.GlobalLog.Log("info", fmt.Sprintf("Starting a file proxy (source: %s) (dest: %s)", conf.ProxySourceFile, conf.ProxyDestFile))

	var obsOverlay *controlpanel.ObsOverlayCode
	go func() error {
		obsOverlay = controlpanel.NewObsOverlayCode(conf.ObsOverlayPort)
		return obsOverlay.Launch()
	}()

	lastLogTime := time.Time{}
	throttleDuration := 1 * time.Second
	throttledLog := func(level string, message string) {
		now := time.Now()
		if now.Sub(lastLogTime) > throttleDuration {
			log.GlobalLog.Log(level, message)
			lastLogTime = now
		}
	}

	// Read the source file periodically
	ticker := time.NewTicker(100 * time.Millisecond)
	lastDisplayCursorX := int(1)
	lastDisplayCursorY := int(1)
	lastEditorCode := []byte{}
	lastRunningCode := []byte{}
	for {
		select {
		case <-ticker.C:
			throttledLog("info", "Tick")

			// Read the source file
			fileData, err := os.ReadFile(conf.ProxySourceFile)
			if err != nil {
				// log but don't exit
				throttledLog("error", fmt.Sprintf("Error reading source file %s: %s", conf.ProxySourceFile, err.Error()))
				continue
			}

			ticStateFromFile, err := tic.MakeTicStateFromExportData(fileData)
			if err != nil {
				// log but don't exit
				throttledLog("error", fmt.Sprintf("could not decode source file: %s", err.Error()))
				continue
			}

			isEditorUpdated := false
			isRunningNewCode := false
			if ticStateFromFile.IsRunning {
				if !bytes.Equal(ticStateFromFile.Code, lastRunningCode) {
					isRunningNewCode = true
				}
				lastRunningCode = ticStateFromFile.Code
			} else {
				hasEditorCodeChanged := !bytes.Equal(ticStateFromFile.Code, lastEditorCode)
				if hasEditorCodeChanged || ticStateFromFile.CursorX != lastDisplayCursorX || ticStateFromFile.CursorY != lastDisplayCursorY {
					isEditorUpdated = true
				}

				// If we're editing, remember this cursor position
				lastDisplayCursorX = ticStateFromFile.CursorX
				lastDisplayCursorY = ticStateFromFile.CursorY
			}

			ticStateToOverlay := tic.State{
				Code:    ticStateFromFile.Code,
				CursorX: lastDisplayCursorX,
				CursorY: lastDisplayCursorY,
			}
			lastEditorCode = ticStateFromFile.Code

			// Send code to OBS overlay
			obsOverlay.SetCode(ticStateToOverlay, "jtruk", isEditorUpdated)

			if isRunningNewCode {
				throttledLog("info", "new code!")

				// Send a version running the last code to TIC
				ticStateToTIC := tic.State{
					IsRunning: false,
					Code:      lastRunningCode,
					CursorX:   1,
					CursorY:   1,
				}

				// Write a version running to the dest file
				dataForTIC, err := ticStateToTIC.MakeDataToImport()
				if err != nil {
					// log but don't exit
					throttledLog("error", fmt.Sprintf("could not encode dest file: %s", err.Error()))
					continue
				}

				err = os.WriteFile(conf.ProxyDestFile, dataForTIC, 0644)
				if err != nil {
					// log but don't exit
					throttledLog("error", fmt.Sprintf("Error writing dest file %s: %s", conf.ProxyDestFile, err.Error()))
					continue
				}

				time.Sleep(200 * time.Millisecond)

				// HACK!
				// Send a version running the last code to TIC
				ticStateToTIC.IsRunning = true

				// Write a version running to the dest file
				dataForTIC, err = ticStateToTIC.MakeDataToImport()
				if err != nil {
					// log but don't exit
					throttledLog("error", fmt.Sprintf("could not encode dest file: %s", err.Error()))
					continue
				}

				err = os.WriteFile(conf.ProxyDestFile, dataForTIC, 0644)
				if err != nil {
					// log but don't exit
					throttledLog("error", fmt.Sprintf("Error writing dest file %s: %s", conf.ProxyDestFile, err.Error()))
					continue
				}
			}

		case <-chUserExitRequest:
			return nil
		}
	}
}
