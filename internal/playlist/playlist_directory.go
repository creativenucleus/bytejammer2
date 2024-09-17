package playlist

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var reLuaFile = regexp.MustCompile(`\.lua$`)

func NewPlaylistDirectory(directory string) (*Playlist, error) {
	p := Playlist{}
	err := p.SyncWithDirectory(directory)
	if err != nil {
		return nil, err
	}
	p.order = ORDER_RANDOM

	return &p, nil
}

// #TODO: deduplicate, and remove those that don't exist
func (p *Playlist) SyncWithDirectory(directory string) error {
	dirEntries, err := os.ReadDir(directory)
	if err != nil {
		return err
	}

	p.items = []PlaylistItem{}
	for _, entry := range dirEntries {
		if entry.IsDir() {
			return fmt.Errorf("directory [%s] contains subdirectories", directory)
		}

		if !reLuaFile.MatchString(entry.Name()) {
			return fmt.Errorf("directory [%s] contains non-lua files", directory)
		}

		path := filepath.Join(directory, entry.Name())
		path, err = filepath.Abs(path)
		if err != nil {
			return err
		}
		p.items = append(p.items, PlaylistItem{
			location: path,
		})
	}

	return nil
}
