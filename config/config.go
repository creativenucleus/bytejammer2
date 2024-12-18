package config

import (
	"encoding/json"
	"fmt"
	"os"
)

var CONFIG Config

type Config struct {
	WorkDir      string             `json:"work_dir"`
	ControlPanel ControlPanelConfig `json:"control_panel"`
	Runnables    RunnablesConfig    `json:"runnables"`
	Jukebox      JukeboxConfig      `json:"jukebox"`
}

type ControlPanelConfig struct {
	Port uint
}

type RunnablesConfig map[string]Runnable

type Runnable struct {
	Filepath string
	Args     []string
}

type JukeboxConfig struct {
	RotatePeriodInSeconds uint `json:"rotate_period_in_seconds"`
}

func LoadGlobal(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading config file [%s]: %w", path, err)
	}

	err = json.Unmarshal(data, &CONFIG)
	if err != nil {
		return fmt.Errorf("error decoding config file [%s]: %w", path, err)
	}

	if CONFIG.Jukebox.RotatePeriodInSeconds == 0 {
		CONFIG.Jukebox.RotatePeriodInSeconds = 15
	}

	return nil
}
