package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Port      string   `yaml:"port"`
	Backends  []string `yaml:"backends"`
	RateLimit struct {
		Capacity    int    `yaml:"capacity"`
		RatePerSec  int    `yaml:"rate_per_sec"`
		StoragePath string `yaml:"storage_path"`
	} `yaml:"ratelimit"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.RateLimit.StoragePath == "" {
		cfg.RateLimit.StoragePath = filepath.Join(".", "ratelimit_data.json")
	}

	return &cfg, nil
}
