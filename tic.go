package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/creativenucleus/bytejammer2/config"
)

type Tic struct {
	MessageBroadcaster
	// The handle to the running TIC
	cmd            *exec.Cmd
	codeImportPath *string
	codeExportPath *string
	// Add latestImport - for server to read
	// Add latestExport - for server to read
}

func NewTic(codeImportPath *string, codeExportPath *string) (*Tic, error) {
	if codeImportPath != nil {
		cleanPath := filepath.Clean(*codeImportPath)
		codeImportPath = &cleanPath
	}

	if codeExportPath != nil {
		cleanPath := filepath.Clean(*codeExportPath)
		codeExportPath = &cleanPath
	}

	tic := &Tic{
		codeImportPath: codeImportPath,
		codeExportPath: codeExportPath,
	}

	return tic, nil
}

func (tic *Tic) messageHandler(message *Message) {
	if message.Type == "tic-state" {
		if tic.codeImportPath != nil {
			ticState, ok := message.Data.(TicState)
			if !ok {
				fmt.Println("Invalid message data type") // #TODO: Raise
				return
			}

			code, err := ticState.MakeDataToImport()
			if err != nil {
				panic(err) // #TODO!
			}

			err = writeCodeToFile(*tic.codeImportPath, string(code))
			if err != nil {
				fmt.Println(err) // #TODO: Raise
				return
			}
		}
	}
}

func writeCodeToFile(codeImportPath string, code string) error {
	return os.WriteFile(codeImportPath, []byte(code), 0600)
}

func (tic *Tic) startMachine() error {
	args := config.CONFIG.Tic80.Args

	if tic.codeImportPath != nil {
		err := EnsurePathExists(filepath.Dir(*tic.codeImportPath), os.ModePerm)
		if err != nil {
			return err
		}

		args = append(args, "--codeimport="+*tic.codeImportPath)
		args = append(args, "--delay=5")
	}

	if tic.codeExportPath != nil {
		err := EnsurePathExists(filepath.Dir(*tic.codeExportPath), os.ModePerm)
		if err != nil {
			return err
		}

		args = append(args, "--codeexport="+*tic.codeExportPath)
	}

	ticExePath := filepath.Clean(config.CONFIG.Tic80.Filepath)
	tic.cmd = exec.Command(ticExePath, args...)
	//	stderr, err := tic.cmd.StderrPipe()
	//	if err != nil {
	//		return err
	//	}

	err := tic.cmd.Start()
	if err != nil {
		return err
	}
	fmt.Printf("Started TIC (pid: %d)\n", tic.cmd.Process.Pid)

	// use goroutine waiting, manage process
	// this is important, otherwise the process becomes in S mode
	go func() {
		//		slurp, _ := io.ReadAll(stderr)
		//		fmt.Printf("%s\n", slurp)

		err = tic.cmd.Wait()
		fmt.Printf("TIC (%d) finished with error: %v", tic.cmd.Process.Pid, err)
		// #TODO: cleanup
	}()

	return nil
	/*
		//	func newTic(slug string, hasImportFile bool, hasExportFile bool, isServer bool /*, broadcaster *NusanLauncher*/ /*) (*Tic, error) {
	tic := Tic{}
	args := []string{
		"--skip",
	}

	fftDevice := os.Getenv("FFTDEVICE")
	if fftDevice != "" {
		fmt.Printf("Sending in FFT Device parameter to TIC: %s\n", fftDevice)
		args = append(args, fmt.Sprintf("--fftdevice=%s", fftDevice))
	}

	exchangefileBasePath := fmt.Sprintf("%s_temp", config.WORK_DIR)
	err := util.EnsurePathExists(exchangefileBasePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	exeBasePath := fmt.Sprintf("%sexecutables", config.WORK_DIR)
	err = util.EnsurePathExists(exeBasePath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	if hasImportFile {
		tic.importFullpath, err = filepath.Abs(fmt.Sprintf("%s/import-%s.lua", exchangefileBasePath, slug))
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("--codeimport=%s", tic.importFullpath))
	}

	if hasExportFile {
		tic.exportFullpath, err = filepath.Abs(fmt.Sprintf("%s/export-%s.lua", exchangefileBasePath, slug))
		if err != nil {
			return nil, err
		}

		args = append(args, fmt.Sprintf("--codeexport=%s", tic.exportFullpath))
	}

	if isServer {
		args = append(args, "--delay=5")
		args = append(args, "--scale=2")
	}

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
