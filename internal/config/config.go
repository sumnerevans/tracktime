package config

import (
	"os"

	"gopkg.in/yaml.v3"

	"github.com/sumnerevans/tracktime/internal/types"
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

type LinearSyncConfig struct {
	DefaultOrg string `yaml:"default_org"`
	APIKey     string `yaml:"api_key"`
}

type SyncConfig struct {
	Enable    bool                `yaml:"enable"`
	GitHub    GitHubSyncConfig    `yaml:"github"`
	GitLab    GitLabSyncConfig    `yaml:"gitlab"`
	SourceHut SourceHutSyncConfig `yaml:"sourcehut"`
	Linear    LinearSyncConfig    `yaml:"linear"`
}

type ReportingConfig struct {
	FullName              string             `yaml:"fullname"`
	ProjectRates          map[string]float64 `yaml:"project_rates"`
	CustomerRates         map[string]float64 `yaml:"customer_rates"`
	CustomerAliases       map[string]string  `yaml:"customer_aliases"`
	CustomerAddresses     map[string]string  `yaml:"customer_addresses"`
	DayWorkedMinThreshold int                `yaml:"day_worked_min_threshold"`
	ReportStatistics      bool               `yaml:"report_statistics"`
	TableFormat           string             `yaml:"table_format"`
}

type Config struct {
	Version   string         `yaml:"version"`
	Directory types.Filename `yaml:"directory"`

	Reporting ReportingConfig `yaml:"reporting"`
	Sync      SyncConfig      `yaml:"sync"`

	// Editor
	Editor     string   `yaml:"editor"`
	EditorArgs []string `yaml:"editor_args"`

	// Typst compiler path (for PDF generation)
	TypstPath string `yaml:"typst_path"`
}

func ReadConfig(f types.Filename) (*Config, error) {
	config := Config{
		Reporting: ReportingConfig{FullName: "<Not Specified>"},
		Directory: types.Filename("$HOME/.tracktime"),
	}
	configData, err := os.ReadFile(f.Expand())
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(configData, &config)
	config.Directory = types.Filename(config.Directory.Expand())
	return &config, err
}
