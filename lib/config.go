package lib

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GitHubSyncConfig struct {
	Username    string `yaml:"username"`
	RootURI     string `yaml:"root_uri"`
	AccessToken string `yaml:"access_token"`
}

type GitLabSyncConfig struct {
	APIRoot string `yaml:"api_root"`
	APIKey  string `yaml:"api_key"`
}

type SourceHutSyncConfig struct {
	APIRoot     string `yaml:"api_root"`
	AccessToken string `yaml:"access_token"`
	Username    string `yaml:"username"`
}

type SyncConfig struct {
	Enable    bool                `yaml:"enable"`
	GitHub    GitHubSyncConfig    `yaml:"github"`
	GitLab    GitLabSyncConfig    `yaml:"gitlab"`
	SourceHut SourceHutSyncConfig `yaml:"sourcehut"`
}

type ReportingConfig struct {
	FullName              string            `yaml:"fullname"`
	ProjectRates          map[string]int    `yaml:"project_rates"`
	CustomerRates         map[string]int    `yaml:"customer_rates"`
	CustomerAliases       map[string]string `yaml:"customer_aliases"`
	CustomerAddresses     map[string]string `yaml:"customer_addresses"`
	DayWorkedMinThreshold int               `yaml:"day_worked_min_threshold"`
	ReportStatistics      bool              `yaml:"report_statistics"`
	// TODO
	// TableFormat           string            `yaml:"table_format"`
}

type Config struct {
	Version   string   `yaml:"version"`
	Directory Filename `yaml:"directory"`

	Reporting ReportingConfig `yaml:"reporting"`
	Sync      SyncConfig      `yaml:"sync"`

	// Editor
	Editor     string   `yaml:"editor"`
	EditorArgs []string `yaml:"editor_args"`
}

func ReadConfig(f Filename) (*Config, error) {
	config := Config{
		Reporting: ReportingConfig{FullName: "<Not Specified>"},
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
