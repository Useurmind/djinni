package config

import (
	"os"
	"path/filepath"

	"github.com/useurmind/djinni/pkg/log"
	"gopkg.in/yaml.v3"
)

func LoadConfig(path string) (*Config, error) {
	log.Info("Loading config...")
	if path == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		path = filepath.Join(cwd, ".djinni.yml")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	ExpandConfigPaths(&cfg)

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	log.Success("Config loaded and validated")
	return &cfg, nil
}
