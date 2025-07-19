package config

import (
	"log/slog"
	"os"

	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		App  `yaml:"app"`
		HTTP `yaml:"http"`
		PGS  `yaml:"postgresql"`
	}

	App struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	HTTP struct {
		Host string `yaml:"host"`
		Port string `yaml:"port"`
	}

	PGS struct {
		Name     string `yaml:"name"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
	}
)

func New() (*Config, error) {
	cfg := &Config{}
	yamlFile, err := os.ReadFile(`configs/main.yaml`)
	if err != nil {
		return &Config{}, err
	}
	if err = yaml.Unmarshal(yamlFile, cfg); err != nil {
		return &Config{}, err
	}
	slog.Info("Конфигурация", "cgf", cfg)
	return cfg, nil
}
