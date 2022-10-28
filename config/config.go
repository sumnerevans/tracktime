package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/sumnerevans/tracktime/lib"
)

type Config struct {
	FullName string `yaml:"fullname"`
}

func ReadConfig(f lib.Filename) (*Config, error) {
	config := Config{
		FullName: "<Not Specified>",
	}
	configData, err := os.ReadFile(f.Expand())
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configData, &config)
	return &config, err
}
