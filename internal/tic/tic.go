package tic

import (
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

func NewMessageTicState(state State) message.Msg {
	return message.Msg{Type: "tic-state", Data: state}
}

// TicManager is a struct that represents a TIC-80 instance
// It can import from a specific path, and export to a specific path
// Currently a TIC instance cannot both import and export
// Implements message.MsgReceiver
type TicManager struct {
	message.MsgSender

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

// Handles "tic-state"
func (tm *TicManager) MsgHandler(msg message.Msg) error {
	switch msg.Type {
	case message.MsgTypeTicState:
		if tm.codeImportPath != nil {
			ticState, ok := msg.Data.(State)
			if !ok {
				fmt.Println("Invalid message data for TicState") // #TODO: Raise
				return nil
			}

			err := tm.writeImportCode(ticState)
			if err != nil {
				fmt.Println(err) // #TODO: Raise
				return nil
			}
		}

	default:
		log.GlobalLog.Send(&message.Msg{Type: message.MsgTypeLog, Data: log.MsgLogData{
			Level:   "error",
			Message: fmt.Sprintf("Unhandled message type [%s] - ignored", msg.Type),
		}})
		return nil
	}

	return nil
}

func (tm *TicManager) StartMachine() error {
	configRunnable := config.CONFIG.Runnables["tic-80-client"]
	args := configRunnable.Args

	if tm.codeImportPath != nil {
		err := files.EnsurePathExists(filepath.Dir(*tm.codeImportPath), os.ModePerm)
		if err != nil {
			return err
		}

		args = append(args, "--codeimport="+*tm.codeImportPath)
		args = append(args, "--delay=5")
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
					log.GlobalLog.Send(&message.Msg{Type: message.MsgTypeLog, Data: log.MsgLogData{
						Level:   "error",
						Message: err.Error(),
					}})
				} else {
					tm.Send(message.Msg{Type: message.MsgTypeTicState, Data: code})
				}
				time.Sleep(5 * time.Second)
			}
		}()
	}

	// use goroutine waiting, manage process
	// this is important, otherwise the process becomes in S mode
	go func() {
		//		slurp, _ := io.ReadAll(stderr)
		//		fmt.Printf("%s\n", slurp)

		err = tm.cmd.Wait()
		log.GlobalLog.Send(&message.Msg{Type: message.MsgTypeLog, Data: log.MsgLogData{Level: "error", Message: fmt.Errorf("TIC (%d) finished with error: %v", tm.cmd.Process.Pid, err).Error()}})

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
		return nil, fmt.Errorf("Tried to export code - but export file is not set up")
	}

	data, err := os.ReadFile(*tm.codeExportPath)
	if err != nil {
		return nil, err
	}

	ts, err := MakeTicStateFromExportData(data)
	if err != nil {
		return nil, err
	}

	return ts, nil
}

func (tm TicManager) writeImportCode(state State) error {
	if tm.codeImportPath == nil {
		return fmt.Errorf("Tried to import code - but import file is not set up")
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
