package files

import (
	"encoding/json"
	"os"
)

func SaveMetaJson(
	fpath string,
	playerName string,
	effectName string,
) error {
	metaData := map[string]string{
		"player_name": playerName,
		"effect_name": effectName,
	}

	data, err := json.MarshalIndent(metaData, "", "\t")
	if err != nil {
		return err
	}

	return os.WriteFile(fpath, data, 0644)
}
