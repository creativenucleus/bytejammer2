package config

import (
	"encoding/json"
	"os"
)

var CONFIG Config

type Config struct {
	WorkDir      string             `json:"work_dir"`
	ControlPanel ControlPanelConfig `json:"control_panel"`
	Runnables    Runnables          `json:"runnables"`
}

type ControlPanelConfig struct {
	Port uint
}

type Runnables map[string]Runnable

type Runnable struct {
	Filepath string
	Args     []string
}

func LoadGlobal(path string) error {
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
