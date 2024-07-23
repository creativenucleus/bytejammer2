package config

import (
	"encoding/json"
	"os"
)

var CONFIG Config

type Config struct {
	WorkDir string      `json:"work_dir"`
	Tic80   Tic80Config `json:"tic-80"`
}

type Tic80Config struct {
	Filepath string
	Args     []string
}

func Load(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &CONFIG)
	if err != nil {
		return err
	}

	return nil
}
