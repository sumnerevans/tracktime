// Package config handles loading and parsing the tracktimerc YAML configuration file.
package config

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	up "go.mau.fi/util/configupgrade"
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

// resolveSecret returns the string as-is unless it ends with "|", in which
// case the prefix is run as a shell command and its trimmed stdout is returned.
func resolveSecret(s string) string {
	if !strings.HasSuffix(s, "|") {
		return s
	}
	out, err := exec.Command("sh", "-c", strings.TrimSuffix(s, "|")).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func ReadConfig(f types.Filename) (*Config, error) {
	configData, _, err := up.Do(f.Expand(), true, Upgrader)
	if err != nil {
		return nil, err
	}

	config := Config{
		Reporting: ReportingConfig{FullName: "<Not Specified>"},
		Directory: types.Filename("$HOME/.tracktime"),
	}
	err = yaml.Unmarshal(configData, &config)
	config.Directory = types.Filename(config.Directory.Expand())
	for i := range config.Logging.Writers {
		config.Logging.Writers[i].Filename = expandEnvXDG(config.Logging.Writers[i].Filename)
	}
	config.GitHub.AccessToken = resolveSecret(config.GitHub.AccessToken)
	config.GitLab.APIKey = resolveSecret(config.GitLab.APIKey)
	config.SourceHut.AccessToken = resolveSecret(config.SourceHut.AccessToken)
	config.Linear.APIKey = resolveSecret(config.Linear.APIKey)
	return &config, err
}
