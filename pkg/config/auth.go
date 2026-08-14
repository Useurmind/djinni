package config

import (
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

func LoadAuthConfig() (*AuthConfig, error) {
	configDir, err := getUserConfigDir()
	if err != nil {
		return nil, err
	}

	authPath := filepath.Join(configDir, "djinni", "config.yaml")

	data, err := os.ReadFile(authPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &AuthConfig{}, nil
		}
		return nil, err
	}

	var cfg AuthConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func getUserConfigDir() (string, error) {
	if dir := os.Getenv("DJINNI_CONFIG_DIR"); dir != "" {
		return dir, nil
	}

	u, err := user.Current()
	if err != nil {
		return "", err
	}

	return filepath.Join(u.HomeDir, ".config"), nil
}

func (c *AuthConfig) Validate() error {
	for _, provider := range c.ModelProviders {
		if err := provider.Validate(); err != nil {
			return err
		}
	}
	return nil
}
