package controlpanel

// Watches a file (potentially deposited by Ticws) and splits into an OBS overlay, and a file for TIC to watch

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/tic"
)

// should either have a client or a host set
type ObsOverlayServerConfig struct {
	ProxyDestFile  string
	PlayerName     string
	ObsOverlayPort uint
}

func ObsOverlayRun(chUserExitRequest <-chan bool, conf ObsOverlayServerConfig, chDataUpdate <-chan []byte) error {
	var obsOverlay *ObsOverlayCode
	go func() error {
		obsOverlay = NewObsOverlayCode(conf.ObsOverlayPort)
		return obsOverlay.Launch()
	}()

	lastLogTime := time.Time{}
	throttleDuration := 5 * time.Second
	throttledLog := func(level string, message string) {
		now := time.Now()
		if now.Sub(lastLogTime) > throttleDuration {
			log.GlobalLog.Log(level, message)
			lastLogTime = now
		}
	}

	// Keep track of last known cursor position and code states
	lastDisplayCursorX := int(1)
	lastDisplayCursorY := int(1)
	lastEditorCode := []byte{}
	lastRunningCode := []byte{}

	for {
		select {
		case fileData := <-chDataUpdate:
			if len(fileData) == 0 {
				// log but don't exit
				throttledLog("error", "Source data file is empty")
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
			obsOverlay.SetCode(ticStateToOverlay, conf.PlayerName, isEditorUpdated)

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
