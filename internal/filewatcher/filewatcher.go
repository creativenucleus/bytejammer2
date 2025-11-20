package filewatcher

import (
	"os"
	"time"
)

func NewFileWatcher(filePath string, d time.Duration, chUserExitRequest <-chan bool) <-chan []byte {
	chFileUpdated := make(chan []byte)

	ticker := time.NewTicker(d)

	go func() {
		for {
			select {
			case <-ticker.C:
				// Read the source file
				fileData, err := os.ReadFile(filePath)
				if err != nil {
					// log but don't exit
					//				throttledLog("error", fmt.Sprintf("Error reading source file %s: %s", conf.ProxySourceFile, err.Error()))
					continue
				}

				chFileUpdated <- fileData

			case <-chUserExitRequest:
				return
			}
		}
	}()

	return chFileUpdated
}
