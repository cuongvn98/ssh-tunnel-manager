package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Server struct {
	ServerAddress string     `yaml:"server_address"`
	Endpoints     []Endpoint `yaml:"endpoints"`
	User          string     `yaml:"user"`
	IdentityFile  string     `yaml:"identity_file"`
	HotkeyCheck   bool       `yaml:"hotkey_check"`
	Timeout       int        `yaml:"timeout"`
}

func (c Server) GetHash() string {
	return fmt.Sprintf("%s:%s:%s:%t", c.ServerAddress, c.User, c.IdentityFile, c.HotkeyCheck)
}

type Endpoint struct {
	RemoteAddress string `yaml:"remote_address"`
	LocalAddress  string `yaml:"local_address"`
}

type Config struct {
	Servers []Server `yaml:"servers"`
}

func LoadConfig(path string) (*Config, error) {
	var config Config

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	err = yaml.NewDecoder(file).Decode(&config)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return &config, nil
}
