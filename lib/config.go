package lib

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	FullName   string   `yaml:"fullname"`
	Directory  Filename `yaml:"directory"`
	Editor     string   `yaml:"editor"`
	EditorArgs []string `yaml:"editor_args"`
}

func ReadConfig(f Filename) (*Config, error) {
	config := Config{
		FullName:  "<Not Specified>",
		Directory: Filename("$HOME/.tracktime"),
	}
	configData, err := os.ReadFile(f.Expand())
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configData, &config)
	config.Directory = Filename(config.Directory.Expand())
	return &config, err
}
