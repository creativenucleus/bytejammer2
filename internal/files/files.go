package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

func EnsurePathExists(path string, perm fs.FileMode) error {
	stat, err := os.Stat(path)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return err
		}

		err := os.MkdirAll(filepath.Clean(path), perm)
		if err != nil {
			return err
		}

		return nil
	}

	if !stat.IsDir() {
		return errors.New("Folder exists, but is not a directory")
	}

	return nil
}

// Not great - but it's a start
func SanitiseFilename(str string) string {
	chars := `ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789`
	pattern := "[^" + chars + "]+"
	r, _ := regexp.Compile(pattern)
	return string(r.ReplaceAll([]byte(str), []byte("")))
}

type FileWatcher struct {
	path             string
	checkFrequency   time.Duration
	chFileDataUpdate chan []byte
}

func NewFileWatcher(path string, checkFrequency time.Duration, chFileDataUpdate chan []byte) (*FileWatcher, error) {
	fw := &FileWatcher{
		path:             path,
		chFileDataUpdate: chFileDataUpdate,
	}

	return fw, nil
}

func (f FileWatcher) Run() {
	ticker := time.NewTicker(f.checkFrequency)

	for {
		<-ticker.C
		data, err := f.Read()
		if err != nil {
			f.chFileDataUpdate <- data
		}
	}
}

func (f FileWatcher) Read() ([]byte, error) {
	return os.ReadFile(f.path)
}
