package tic

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strconv"
	"time"

	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
)

type Replayer struct {
	message.MsgPropagator
	zipFile string
}

func NewReplayer(zipFile string) (*Replayer, error) {
	r := Replayer{
		zipFile: zipFile,
	}

	return &r, nil
}

func (r *Replayer) Run(chUserExitRequest <-chan bool) error {
	log.GlobalLog.Log("info", "Replayer starting")

	archive, err := zip.OpenReader(r.zipFile)
	if err != nil {
		return err
	}
	defer archive.Close()

	// sort the files in order of timestamp
	sort.Slice(archive.File, func(i, j int) bool { return archive.File[i].Name < archive.File[j].Name })

	reFilename := regexp.MustCompile(`^snap-(\d{1,20})$`)
	prevStamp := 0
	for _, f := range archive.File {
		if f.FileInfo().IsDir() {
			return err
		}

		res := reFilename.FindStringSubmatch(f.Name)
		if len(res) != 2 {
			return err
		}

		stamp, err := strconv.Atoi(res[1])
		if err != nil {
			return err
		}

		if prevStamp != 0 {
			stampWaitMS := (int)(stamp - prevStamp)
			timer := time.NewTimer(time.Duration(stampWaitMS) * time.Nanosecond)
			select {
			case <-timer.C:
			case <-chUserExitRequest:
				return nil // Allow the user to exit during the sleep
			}
		}
		prevStamp = stamp

		fileInArchive, err := f.Open()
		if err != nil {
			return err
		}

		data, err := io.ReadAll(fileInArchive)
		if err != nil {
			return err
		}
		fileInArchive.Close()

		state, err := MakeTicStateFromExportData(data)
		if err != nil {
			return err
		}

		msgData, err := json.Marshal(state)
		if err != nil {
			return err
		}

		fmt.Printf("Sending: %s\n", f.Name)
		r.Propagate(message.MsgTypeTicState, msgData)
	}

	return nil
}
