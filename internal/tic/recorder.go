package tic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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
	log.GlobalLog.Log("info", "Recorder starting")

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
