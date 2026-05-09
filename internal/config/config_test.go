package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sumnerevans/tracktime/internal/types"
)

func TestReadConfig(t *testing.T) {
	t.Run("minimal config file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "tracktimerc")
		require.NoError(t, os.WriteFile(configFile, []byte("directory: $HOME/.tracktime\n"), 0644))

		cfg, err := ReadConfig(types.Filename(configFile))
		require.NoError(t, err)

		homeDir, _ := os.UserHomeDir()
		assert.Equal(t, types.Filename(filepath.Join(homeDir, ".tracktime")), cfg.Directory)
		assert.Equal(t, "<Not Specified>", cfg.Reporting.FullName)
		// Base config defaults are applied by the upgrader.
		assert.Equal(t, 120, cfg.Reporting.DayWorkedMinThreshold)
		assert.True(t, cfg.Reporting.ReportStatistics)
		assert.Equal(t, 30, cfg.ItemCacheTTLDays)
		assert.Equal(t, "https://gitlab.com/api/v4/", cfg.GitLab.APIRoot)
		assert.NotNil(t, cfg.Reporting.ProjectRates)
		assert.NotNil(t, cfg.Reporting.CustomerRates)
	})

	t.Run("config file with environment variable expansion", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "tracktimerc")

		t.Setenv("TEST_TRACKTIME_DIR", "/custom/tracktime/path")

		configContent := "directory: \"$TEST_TRACKTIME_DIR\"\n"
		require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

		cfg, err := ReadConfig(types.Filename(configFile))
		require.NoError(t, err)

		assert.Equal(t, types.Filename("/custom/tracktime/path"), cfg.Directory)
		assert.Equal(t, "<Not Specified>", cfg.Reporting.FullName)
	})

	t.Run("migrates Python flat format to Go nested format", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "tracktimerc")

		configContent := `fullname: Test User
sync_time: true
day_worked_min_threshold: 60
report_statistics: false
project_rates:
  myproject: 100.0
customer_aliases:
  ACME: ACME Corp
editor: vim
editor_args: --noplugin,--clean
`
		require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

		cfg, err := ReadConfig(types.Filename(configFile))
		require.NoError(t, err)

		assert.Equal(t, "Test User", cfg.Reporting.FullName)
		assert.True(t, cfg.Sync.Enable)
		assert.Equal(t, 60, cfg.Reporting.DayWorkedMinThreshold)
		assert.False(t, cfg.Reporting.ReportStatistics)
		assert.Equal(t, map[string]float64{"myproject": 100.0}, cfg.Reporting.ProjectRates)
		assert.Equal(t, map[string]string{"ACME": "ACME Corp"}, cfg.Reporting.CustomerAliases)
		assert.Equal(t, "vim", cfg.Editor)
		assert.Equal(t, []string{"--noplugin", "--clean"}, cfg.EditorArgs)
	})

	t.Run("Go-format config is unchanged by upgrader", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "tracktimerc")

		configContent := `reporting:
  fullname: Go User
  day_worked_min_threshold: 90
  report_statistics: false
  project_rates:
    proj: 75.0
sync:
  enable: true
editor_args:
  - --wait
`
		require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

		cfg, err := ReadConfig(types.Filename(configFile))
		require.NoError(t, err)

		assert.Equal(t, "Go User", cfg.Reporting.FullName)
		assert.Equal(t, 90, cfg.Reporting.DayWorkedMinThreshold)
		assert.False(t, cfg.Reporting.ReportStatistics)
		assert.Equal(t, map[string]float64{"proj": 75.0}, cfg.Reporting.ProjectRates)
		assert.True(t, cfg.Sync.Enable)
		assert.Equal(t, []string{"--wait"}, cfg.EditorArgs)
	})
}
