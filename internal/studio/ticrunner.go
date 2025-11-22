package studio

// TODO: Possibly more flexible if not in the studio package?

// Watches a file (potentially deposited by Ticws) and sends to a file for TIC to watch

import (
	"fmt"
	"os"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/tic"
)

func TicRunner(
	chUserExitRequest <-chan bool,
	chDataUpdate <-chan []byte,
	fileToWritePath string,
) error {
	lastLogTime := time.Time{}
	throttleDuration := 5 * time.Second
	throttledLog := func(level string, message string) {
		now := time.Now()
		if now.Sub(lastLogTime) > throttleDuration {
			log.GlobalLog.Log(level, message)
			lastLogTime = now
		}
	}

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

			// Write a version running to the dest file
			dataForTIC, err := ticStateFromFile.MakeDataToImport()
			if err != nil {
				// log but don't exit
				throttledLog("error", fmt.Sprintf("could not encode dest file: %s", err.Error()))
				continue
			}

			err = os.WriteFile(fileToWritePath, dataForTIC, 0644)
			if err != nil {
				// log but don't exit
				throttledLog("error", fmt.Sprintf("Error writing dest file %s: %s", fileToWritePath, err.Error()))
				continue
			}

			time.Sleep(200 * time.Millisecond)

		case <-chUserExitRequest:
			return nil
		}
	}
}
