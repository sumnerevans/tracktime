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
		configContent := `version: "1.0"
`
		require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

		cfg, err := ReadConfig(types.Filename(configFile))
		require.NoError(t, err)

		homeDir, _ := os.UserHomeDir()
		expected := &Config{
			Version:   "1.0",
			Directory: types.Filename(filepath.Join(homeDir, ".tracktime")),
			Reporting: ReportingConfig{
				FullName: "<Not Specified>",
			},
		}
		assert.EqualValues(t, expected, cfg)
	})

	t.Run("config file with environment variable expansion", func(t *testing.T) {
		tmpDir := t.TempDir()
		configFile := filepath.Join(tmpDir, "tracktimerc")

		// Set a test environment variable
		os.Setenv("TEST_TRACKTIME_DIR", "/custom/tracktime/path")
		defer os.Unsetenv("TEST_TRACKTIME_DIR")

		configContent := `version: "1.0"
directory: "$TEST_TRACKTIME_DIR"
`
		require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0644))

		cfg, err := ReadConfig(types.Filename(configFile))
		require.NoError(t, err)

		expected := &Config{
			Version:   "1.0",
			Directory: types.Filename("/custom/tracktime/path"),
			Reporting: ReportingConfig{
				FullName: "<Not Specified>",
			},
		}
		assert.EqualValues(t, expected, cfg)
	})
}
