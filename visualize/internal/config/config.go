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

type Scheme struct {
	Active       bool              `yaml:"active" json:"active"`
	Background   string            `yaml:"background" json:"background"`
	Foreground   string            `yaml:"foreground" json:"foreground"`
	DefaultColor string            `yaml:"default_color" json:"default_color"`
	Messages     map[string]string `yaml:"messages" json:"messages"`
}

type Config struct {
	Listen      string            `yaml:"listen" json:"listen"`
	BufferSize  int               `yaml:"buffer_size" json:"buffer_size"`
	Mode        string            `yaml:"mode" json:"mode"`
	Orientation string            `yaml:"orientation" json:"orientation"`
	Resolution  string            `yaml:"resolution" json:"resolution"`
	Schemes     map[string]Scheme `yaml:"schemes" json:"schemes"`
	Lanes       []Lane            `yaml:"lanes" json:"lanes"`
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
