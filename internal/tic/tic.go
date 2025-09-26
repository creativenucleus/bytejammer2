package tic

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/creativenucleus/bytejammer2/config"
	"github.com/creativenucleus/bytejammer2/internal/files"
	"github.com/creativenucleus/bytejammer2/internal/log"
	"github.com/creativenucleus/bytejammer2/internal/message"
)

type MsgTicStateData struct {
	State
}

func NewMessageTicState(state State) (*message.Msg, error) {
	/*	data, err := json.Marshal(state)
		if err != nil {
			return nil, err
		}
	*/
	data := map[string]any{
		"state": state,
	}

	return &message.Msg{Type: "tic-state", Data: data}, nil
}

type MsgTicSnapshotData struct {
	PlayerName string
	EffectName string
	Code       []byte
}

// TicManager is a struct that represents a TIC-80 instance
// It can import from a specific path, and export to a specific path
// Currently a TIC instance cannot both import and export
// Implements message.MsgReceiver
type TicManager struct {
	message.MsgPropagator

	// The handle to the running TIC
	cmd            *exec.Cmd
	codeImportPath *string
	codeExportPath *string
	// Add latestImport - for server to read
	// Add latestExport - for server to read
}

// NewTic creates a new TicManager
func NewTicManager(codeImportPath *string, codeExportPath *string /*, broadcaster *NusanLauncher*/) (*TicManager, error) {
	if codeImportPath != nil {
		cleanPath := filepath.Clean(*codeImportPath)
		codeImportPath = &cleanPath
	}

	if codeExportPath != nil {
		cleanPath := filepath.Clean(*codeExportPath)
		codeExportPath = &cleanPath
	}

	tic := &TicManager{
		codeImportPath: codeImportPath,
		codeExportPath: codeExportPath,
	}

	return tic, nil
}

func (tm *TicManager) GetState() (*State, error) {
	if tm.codeExportPath == nil {
		return nil, fmt.Errorf("tried to export code - but export file is not set up")
	}

	return tm.readExportCode()
}

func (tm *TicManager) SetState(state State) error {
	if tm.codeImportPath == nil {
		return fmt.Errorf("tried to import code - but import file is not set up")
	}

	return tm.writeImportCode(state)
}

// Handles "tic-state"
func (tm *TicManager) MsgHandler(msgType message.MsgType, msgData []byte) error {
	switch msgType {
	case message.MsgTypeTicState:
		if tm.codeImportPath != nil {
			var ticState State
			err := json.Unmarshal(msgData, &ticState)
			if err != nil {
				fmt.Println("Invalid message data for TicState") // #TODO: Raise
				return err
			}

			err = tm.writeImportCode(ticState)
			if err != nil {
				fmt.Println(err) // #TODO: Raise
				return nil
			}
		}

	default:
		log.GlobalLog.Log("error", fmt.Sprintf("Unhandled message type [%s] - ignored", msgType))
		return nil
	}

	return nil
}

func (tm *TicManager) StartMachine(machineConfig string) error {
	configRunnable, ok := config.CONFIG.Runnables[machineConfig]
	if !ok {
		return fmt.Errorf("could not find runnable setting %s in config.json", machineConfig)
	}
	args := configRunnable.Args

	if tm.codeImportPath != nil {
		err := files.EnsurePathExists(filepath.Dir(*tm.codeImportPath), os.ModePerm)
		if err != nil {
			return err
		}

		args = append(args, "--codeimport="+*tm.codeImportPath)
		args = append(args, "--delay=1")
	}

	if tm.codeExportPath != nil {
		err := files.EnsurePathExists(filepath.Dir(*tm.codeExportPath), os.ModePerm)
		if err != nil {
			return err
		}

		args = append(args, "--codeexport="+*tm.codeExportPath)
	}

	ticExePath := filepath.Clean(configRunnable.Filepath)
	tm.cmd = exec.Command(ticExePath, args...)
	//	stderr, err := tic.cmd.StderrPipe()
	//	if err != nil {
	//		return err
	//	}

	err := tm.cmd.Start()
	if err != nil {
		return err
	}

	if tm.codeExportPath != nil {
		go func() {
			for {
				code, err := tm.readExportCode()
				if err != nil {
					log.GlobalLog.Log("error", err.Error())
				} else {
					encCode, err := json.Marshal(code)
					if err != nil {
						log.GlobalLog.Log("error", err.Error())
					}

					tm.Propagate(message.MsgTypeTicState, encCode)
				}
				time.Sleep(500 * time.Millisecond)
			}
		}()
	}

	// use goroutine waiting, manage process
	// this is important, otherwise the process becomes in S mode
	go func() {
		//		slurp, _ := io.ReadAll(stderr)
		//		fmt.Printf("%s\n", slurp)

		err = tm.cmd.Wait()
		log.GlobalLog.Log("error", fmt.Errorf("TIC (%d) finished with error: %v", tm.cmd.Process.Pid, err).Error())

		// #TODO: cleanup
	}()

	return nil
	/*

		//	if broadcaster == nil {
		fmt.Printf("Running TIC-80 version [%s]\n", embed.Tic80version)

		// #TODO: multiversion
		tic.ticFilename = filepath.Clean(fmt.Sprintf("%s/tic80-%s.exe", exeBasePath, embed.Tic80version))
		_, err = os.Stat(tic.ticFilename)
		if err != nil {
			if !os.IsNotExist(err) { // An error we won't handle
				return nil, err
			} else { // File doesn't exist - try creating it...
				err = os.WriteFile(tic.ticFilename, embed.Tic80exe, 0700)
				if err != nil {
					return nil, err
				}
			}
		}

		tic.cmd = exec.Command(tic.ticFilename, args...)
		err = tic.cmd.Start()
		if err != nil {
			return nil, err
		}

		fmt.Printf("Started TIC (pid: %d)\n", tic.cmd.Process.Pid)

		// use goroutine waiting, manage process
		// this is important, otherwise the process becomes in S mode
		go func() {
			err = tic.cmd.Wait()
			fmt.Printf("TIC (%d) finished with error: %v", tic.cmd.Process.Pid, err)
			// #TODO: cleanup
		}()
	*/
	/*
		} else {
			fmt.Printf("Running broadcast TIC-80 version\n")

			(*broadcaster.ch) <- fmt.Sprintf("--codeimport=%s", filepath.Clean(tic.importFullpath))
		}
	*/
	/*
		return &tic, nil
	*/
}

func (tm TicManager) readExportCode() (*State, error) {
	if tm.codeExportPath == nil {
		return nil, fmt.Errorf("tried to export code - but export file is not set up")
	}

	data, err := os.ReadFile(*tm.codeExportPath)
	if err != nil {
		return nil, err
	}

	state, err := MakeTicStateFromExportData(data)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func (tm TicManager) writeImportCode(state State) error {
	if tm.codeImportPath == nil {
		return fmt.Errorf("tried to import code - but import file is not set up")
	}

	data, err := state.MakeDataToImport()
	if err != nil {
		return err
	}

	err = os.WriteFile(*tm.codeImportPath, data, 0600)
	if err != nil {
		return err
	}

	return nil
}
