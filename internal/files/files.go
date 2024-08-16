package files

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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
