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

func LoadMetaJson(filePath string) (map[string]string, error) {
	metaJson, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var metaData map[string]string
	err = json.Unmarshal(metaJson, &metaData)
	if err != nil {
		return nil, err
	}

	return metaData, nil
}
