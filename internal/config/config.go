package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	FPSUnlock      bool    `yaml:"fps_unlock"`
	FPS            int     `yaml:"fps"`
	Resolution     bool    `yaml:"resolution"`
	Width          int     `yaml:"width"`
	Height         int     `yaml:"height"`
	FOV            bool    `yaml:"fov"`
	FOVValue       float64 `yaml:"fov_value"`
	CameraReset    bool    `yaml:"camera_reset"`
	AutoLoot       bool    `yaml:"auto_loot"`
	Dragonrot      bool    `yaml:"dragonrot"`
	DeathPenalty   bool    `yaml:"death_penalty"`
	GameSpeed      bool    `yaml:"game_speed"`
	GameSpeedVal   float64 `yaml:"game_speed_value"`
	PlayerSpeed    bool    `yaml:"player_speed"`
	PlayerSpeedVal float64 `yaml:"player_speed_value"`
}

func DefaultConfig() *Config {
	return &Config{
		FPSUnlock:      true,
		FPS:            144,
		Resolution:     false,
		Width:          1920,
		Height:         1080,
		FOV:            false,
		FOVValue:       1.0,
		CameraReset:    false,
		AutoLoot:       false,
		Dragonrot:      false,
		DeathPenalty:   false,
		GameSpeed:      false,
		GameSpeedVal:   1.0,
		PlayerSpeed:    false,
		PlayerSpeedVal: 1.0,
	}
}

func getConfigPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	appConfigDir := filepath.Join(configDir, "sekiro-tweaker")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appConfigDir, "config.yaml"), nil
}

func Load() *Config {
	configPath, err := getConfigPath()
	if err != nil {
		return DefaultConfig()
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig()
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return DefaultConfig()
	}

	return &config
}

func (c *Config) Save() {
	configPath, err := getConfigPath()
	if err != nil {
		return
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return
	}

	_ = os.WriteFile(configPath, data, 0644)
}
