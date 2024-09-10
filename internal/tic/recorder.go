package tic

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
)

type Recorder struct {
	basePath  string
	lastState State
}

func NewRecorder(basePath string) (*Recorder, error) {
	r := Recorder{
		basePath: basePath,
	}

	return &r, nil
}

func (r *Recorder) MsgHandler(msgType message.MsgType, msgData []byte) error {
	switch msgType {
	case message.MsgTypeTicState:
		var ticState State
		err := json.Unmarshal(msgData, &ticState)
		if err != nil {
			return err
		}

		if ticState.IsEqual(r.lastState) {
			return nil // no change
		}

		data, err := ticState.MakeDataToImport()
		if err != nil {
			return err
		}

		t := time.Now()
		filename := filepath.Join(r.basePath, fmt.Sprintf("snap-%d", t.UnixNano()))

		err = os.WriteFile(filename, data, 0644)
		if err != nil {
			return err
		}

		r.lastState = ticState
	}

	return nil
}

// Put everything into a zip file...
func (r *Recorder) Close() error {
	log.GlobalLog.Log("info", "Recorder closing")

	t := time.Now()
	zipFilename := filepath.Join(r.basePath, fmt.Sprintf("recorder-%s.zip", t.Format("2006-01-02-15-04-05")))
	archive, err := os.Create(zipFilename)
	if err != nil {
		return err
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	fileInfo, err := os.ReadDir(r.basePath)
	if err != nil {
		return err
	}

	var filesToRemove []string
	reFilename := regexp.MustCompile(`^snap-(\d{1,20})$`)
	for _, file := range fileInfo {
		if file.IsDir() {
			continue
		}

		if !reFilename.MatchString(file.Name()) {
			continue
		}

		snapFilepath := filepath.Join(r.basePath, file.Name())

		// Add file wrapper (to handle defer)
		err = func(zw *zip.Writer, filepath string, filename string) error {
			snapFP, err := os.Open(snapFilepath)
			if err != nil {
				return err
			}
			defer snapFP.Close()

			fileWriter, err := zipWriter.Create(filename)
			if err != nil {
				return err
			}

			_, err = io.Copy(fileWriter, snapFP)
			if err != nil {
				return err
			}

			return nil
		}(zipWriter, snapFilepath, file.Name())
		if err != nil {
			return err
		}

		filesToRemove = append(filesToRemove, snapFilepath)
	}

	err = zipWriter.Close()
	if err != nil {
		return err
	}

	// Delete all the snap files
	for _, filename := range filesToRemove {
		fmt.Printf("Removing: %s\n", filename)
		err = os.Remove(filename)
		if err != nil {
			return err
		}
	}

	return nil
}
