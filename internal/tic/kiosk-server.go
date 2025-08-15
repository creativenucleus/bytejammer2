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

		fnameBase := fmt.Sprintf("%s-%s-%s", timeNow.Format("20060102150405"), files.SanitiseFilename(snapshotData.PlayerName), files.SanitiseFilename(snapshotData.EffectName))
		fpathLua := fmt.Sprintf("%s/%s.lua", ks.directory, fnameBase)
		fpathMetaJson := fmt.Sprintf("%s.meta.json", fpathLua)

		// TODO: Check for breakout
		cleanPath, err := filepath.Abs(fpathLua)
		if err != nil {
			return err
		}

		err = os.WriteFile(cleanPath, snapshotData.Code, 0644)
		if err != nil {
			return err
		}

		err = files.SaveMetaJson(fpathMetaJson, snapshotData.PlayerName, snapshotData.EffectName)
		if err != nil {
			return err
		}

		log.GlobalLog.Log("info", fmt.Sprintf("Received and saved TIC Snapshot to kiosk directory: %s", fnameBase))
	}

	return nil
}
