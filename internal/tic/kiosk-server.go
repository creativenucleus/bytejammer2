package tic

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"encoding/json"

	"github.com/creativenucleus/bytejammer2/internal/files"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
)

type KioskServer struct {
	directory string
}

func NewKioskServer(dir string) *KioskServer {
	return &KioskServer{
		directory: dir,
	}
}

func (ks *KioskServer) MsgHandler(msgType message.MsgType, msgData []byte) error {
	switch msgType {
	case message.MsgTypeTicSnapshot:
		var snapshotData MsgTicSnapshotData
		err := json.Unmarshal(msgData, &snapshotData)
		if err != nil {
			return err
		}

		timeNow := time.Now()

		fname := fmt.Sprintf("%s-%s.lua", timeNow.Format("20060102150405"), files.SanitiseFilename(snapshotData.DisplayName))
		fpath := fmt.Sprintf("%s/%s", ks.directory, fname)
		// TODO: Check for breakout
		cleanPath, err := filepath.Abs(fpath)
		if err != nil {
			return err
		}

		err = os.WriteFile(cleanPath, snapshotData.Code, 0644)
		if err != nil {
			return err
		}

		log.GlobalLog.Log("info", fmt.Sprintf("Received and saved TIC Snapshot to kiosk directory: %s", fname))
	}

	return nil
}
