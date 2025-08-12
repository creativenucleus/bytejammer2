package playlist

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
)

var reLuaFile = regexp.MustCompile(`\.lua$`)

func NewPlaylistDirectory(directory string) (*Playlist, error) {
	p := Playlist{}
	_, err := p.SyncWithDirectory(directory)
	if err != nil {
		return nil, err
	}
	p.order = ORDER_ITERATE

	return &p, nil
}

// #TODO: deduplicate, and remove those that don't exist
// Returns (isUpdated, error)
// Add new items to the front of the playlist (so we can play them immediately)
func (p *Playlist) SyncWithDirectory(directory string) (bool, error) {
	dirEntries, err := os.ReadDir(directory)
	if err != nil {
		return false, err
	}

	isUpdated := false
	itemsInDirectory := []string{}
	for _, entry := range dirEntries {
		if entry.IsDir() {
			return false, fmt.Errorf("directory [%s] contains subdirectories", directory)
		}

		if !reLuaFile.MatchString(entry.Name()) {
			// Skip this one...
			continue
		}

		path := filepath.Join(directory, entry.Name())
		path, err = filepath.Abs(path)
		if err != nil {
			return false, err
		}
		itemsInDirectory = append(itemsInDirectory, path)
	}

	// Iterate through each item in our previous list
	// If it's in our new list then append it in the same order, and remove from our appending list
	// #TODO: fix slice reordering if needed
	newItems := []PlaylistItem{}
	for _, item := range p.items {
		itemIndex := slices.Index(itemsInDirectory, item.location)
		if itemIndex == -1 { // not found
			isUpdated = true
		} else {
			newItems = append(newItems, item)
			itemsInDirectory = slices.Delete(itemsInDirectory, itemIndex, itemIndex+1)
		}
	}

	// Add all items that are in our new list but not found in our old list
	for _, itemPath := range itemsInDirectory {
		newItems = slices.Insert(newItems, 0, PlaylistItem{
			location: itemPath,
		})
		isUpdated = true
	}

	// If updated, reset to the first in the list
	if isUpdated {
		p.previous = -1
	}

	p.items = newItems
	return isUpdated, nil
}
