// Package config reads visualize.yml.
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Lane struct {
	Name string `yaml:"name" json:"name"`
	Path string `yaml:"path" json:"path"`
}

type Config struct {
	Listen     string `yaml:"listen" json:"listen"`
	BufferSize int    `yaml:"buffer_size" json:"buffer_size"`
	Scheme     string `yaml:"scheme" json:"scheme"`
	Lanes      []Lane `yaml:"lanes" json:"lanes"`
}

func Load(path string) (*Config, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(body, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
