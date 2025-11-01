package synchroniser

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
)

func TestLinearSynchroniser_GetFormattedTaskID(t *testing.T) {
	sync := &LinearSynchroniser{}
	sync.Init(config.SyncConfig{
		Linear: config.LinearSyncConfig{
			DefaultOrg: "myorg",
			APIKey:     "test-key",
		},
	})

	tests := []struct {
		name     string
		entry    *timeentry.TimeEntry
		expected string
	}{
		{
			name: "valid linear entry",
			entry: &timeentry.TimeEntry{
				Type:    "linear",
				Project: "ENG",
				TaskID:  "123",
			},
			expected: "ENG-123",
		},
		{
			name: "different project prefix",
			entry: &timeentry.TimeEntry{
				Type:    "linear",
				Project: "PROD",
				TaskID:  "456",
			},
			expected: "PROD-456",
		},
		{
			name: "wrong type returns empty",
			entry: &timeentry.TimeEntry{
				Type:    "github",
				Project: "ENG",
				TaskID:  "123",
			},
			expected: "",
		},
		{
			name: "missing project returns empty",
			entry: &timeentry.TimeEntry{
				Type:   "linear",
				TaskID: "123",
			},
			expected: "",
		},
		{
			name: "missing taskID returns empty",
			entry: &timeentry.TimeEntry{
				Type:    "linear",
				Project: "ENG",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sync.GetFormattedTaskID(tt.entry)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLinearSynchroniser_GetTaskLink(t *testing.T) {
	sync := &LinearSynchroniser{}
	sync.Init(config.SyncConfig{
		Linear: config.LinearSyncConfig{
			DefaultOrg: "myorg",
			APIKey:     "test-key",
		},
	})

	tests := []struct {
		name     string
		entry    *timeentry.TimeEntry
		expected string
	}{
		{
			name: "valid linear entry",
			entry: &timeentry.TimeEntry{
				Type:    "linear",
				Project: "ENG",
				TaskID:  "123",
			},
			expected: "https://linear.app/myorg/issue/ENG-123",
		},
		{
			name: "different issue ID",
			entry: &timeentry.TimeEntry{
				Type:    "linear",
				Project: "PROD",
				TaskID:  "456",
			},
			expected: "https://linear.app/myorg/issue/PROD-456",
		},
		{
			name: "wrong type returns empty",
			entry: &timeentry.TimeEntry{
				Type:    "github",
				Project: "ENG",
				TaskID:  "123",
			},
			expected: "",
		},
		{
			name: "missing project returns empty",
			entry: &timeentry.TimeEntry{
				Type:   "linear",
				TaskID: "123",
			},
			expected: "",
		},
		{
			name: "missing taskID returns empty",
			entry: &timeentry.TimeEntry{
				Type:    "linear",
				Project: "ENG",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sync.GetTaskLink(tt.entry)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLinearSynchroniser_GetTaskLink_NoDefaultOrg(t *testing.T) {
	sync := &LinearSynchroniser{}
	sync.Init(config.SyncConfig{
		Linear: config.LinearSyncConfig{
			APIKey: "test-key",
			// No DefaultOrg set
		},
	})

	entry := &timeentry.TimeEntry{
		Type:    "linear",
		Project: "ENG",
		TaskID:  "123",
	}

	result := sync.GetTaskLink(entry)
	assert.Equal(t, "", result, "Should return empty string when DefaultOrg is not configured")
}
