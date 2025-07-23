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
		LOG  `yaml:"logs"`
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

	LOG struct {
		LogLevel int `yaml:"loglevel"`
	}
)

func New() (*Config, error) {
	cfg := &Config{}

	yamlFile, err := os.ReadFile(`/app/config.yaml`)
	//yamlFile, err := os.ReadFile(`.env/config.yaml`)
	if err != nil {
		slog.Error(err.Error())
		return &Config{}, err
	}
	if err = yaml.Unmarshal(yamlFile, cfg); err != nil {
		slog.Error(err.Error())
		return &Config{}, err
	}
	slog.Debug("Конфигурация", "cgf", cfg)
	return cfg, nil
}
