// Package config handles loading and parsing the tracktimerc YAML configuration file.
package config

import (
	"os"
	"path/filepath"
	"time"

	"go.mau.fi/zeroconfig"
	"gopkg.in/yaml.v3"

	"github.com/sumnerevans/tracktime/internal/types"
)

type GitHubConfig struct {
	Username    string `yaml:"username"`
	RootURI     string `yaml:"root_uri"`
	AccessToken string `yaml:"access_token"`
}

type GitLabConfig struct {
	APIRoot string `yaml:"api_root"`
	APIKey  string `yaml:"api_key"`
}

type SourceHutConfig struct {
	APIRoot     string `yaml:"api_root"`
	AccessToken string `yaml:"access_token"`
	Username    string `yaml:"username"`
}

type LinearConfig struct {
	DefaultOrg string `yaml:"default_org"`
	APIKey     string `yaml:"api_key"`
}

type SyncConfig struct {
	Enable bool `yaml:"enable"`
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

	Reporting        ReportingConfig `yaml:"reporting"`
	Sync             SyncConfig      `yaml:"sync"`
	GitHub           GitHubConfig    `yaml:"github"`
	GitLab           GitLabConfig    `yaml:"gitlab"`
	SourceHut        SourceHutConfig `yaml:"sourcehut"`
	Linear           LinearConfig    `yaml:"linear"`
	ItemCacheTTLDays int             `yaml:"item_cache_ttl_days"`

	// Editor
	Editor     string   `yaml:"editor"`
	EditorArgs []string `yaml:"editor_args"`

	// Typst compiler path (for PDF generation)
	TypstPath string `yaml:"typst_path"`

	Logging zeroconfig.Config `yaml:"logging"`
}

// CacheTTL returns the configured item description cache TTL, defaulting to 30 days.
func (c *Config) CacheTTL() time.Duration {
	if c.ItemCacheTTLDays <= 0 {
		return 30 * 24 * time.Hour
	}
	return time.Duration(c.ItemCacheTTLDays) * 24 * time.Hour
}

// expandEnvXDG expands environment variables in s, providing the XDG base
// directory defaults for variables that are unset in the environment.
func expandEnvXDG(s string) string {
	home := os.Getenv("HOME")
	xdgDefaults := map[string]string{
		"XDG_STATE_HOME":  filepath.Join(home, ".local", "state"),
		"XDG_DATA_HOME":   filepath.Join(home, ".local", "share"),
		"XDG_CACHE_HOME":  filepath.Join(home, ".cache"),
		"XDG_CONFIG_HOME": filepath.Join(home, ".config"),
	}
	return os.Expand(s, func(key string) string {
		if val := os.Getenv(key); val != "" {
			return val
		}
		if def, ok := xdgDefaults[key]; ok {
			return def
		}
		return ""
	})
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
	for i := range config.Logging.Writers {
		config.Logging.Writers[i].Filename = expandEnvXDG(config.Logging.Writers[i].Filename)
	}
	return &config, err
}
