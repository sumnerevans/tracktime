package resolver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sumnerevans/tracktime/internal/config"
	"github.com/sumnerevans/tracktime/internal/timeentry"
)

func TestLinearResolver_GetFormattedTaskID(t *testing.T) {
	r := &LinearResolver{}
	r.Init(&config.Config{
		Linear: config.LinearConfig{
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
			assert.Equal(t, tt.expected, r.GetFormattedTaskID(tt.entry))
		})
	}
}

func TestLinearResolver_GetTaskLink(t *testing.T) {
	r := &LinearResolver{}
	r.Init(&config.Config{
		Linear: config.LinearConfig{
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
			assert.Equal(t, tt.expected, r.GetTaskLink(tt.entry))
		})
	}
}

func TestLinearResolver_GetTaskLink_NoDefaultOrg(t *testing.T) {
	r := &LinearResolver{}
	r.Init(&config.Config{
		Linear: config.LinearConfig{
			APIKey: "test-key",
		},
	})

	entry := &timeentry.TimeEntry{
		Type:    "linear",
		Project: "ENG",
		TaskID:  "123",
	}

	assert.Equal(t, "", r.GetTaskLink(entry), "Should return empty string when DefaultOrg is not configured")
}

func TestLinearResolver_FetchDescription_GuardClauses(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		cfg    config.LinearConfig
		entry  *timeentry.TimeEntry
		expect string
	}{
		{
			name: "wrong type returns empty",
			cfg:  config.LinearConfig{DefaultOrg: "myorg", APIKey: "test-key"},
			entry: &timeentry.TimeEntry{
				Type:    "github",
				Project: "ENG",
				TaskID:  "123",
			},
			expect: "",
		},
		{
			name: "missing taskid returns empty",
			cfg:  config.LinearConfig{DefaultOrg: "myorg", APIKey: "test-key"},
			entry: &timeentry.TimeEntry{
				Type:    "linear",
				Project: "ENG",
			},
			expect: "",
		},
		{
			name: "missing api key returns empty",
			cfg:  config.LinearConfig{DefaultOrg: "myorg"},
			entry: &timeentry.TimeEntry{
				Type:    "linear",
				Project: "ENG",
				TaskID:  "123",
			},
			expect: "",
		},
		{
			name: "missing default org returns empty",
			cfg:  config.LinearConfig{APIKey: "test-key"},
			entry: &timeentry.TimeEntry{
				Type:    "linear",
				Project: "ENG",
				TaskID:  "123",
			},
			expect: "",
		},
		{
			name: "missing project returns empty (no formatted task id)",
			cfg:  config.LinearConfig{DefaultOrg: "myorg", APIKey: "test-key"},
			entry: &timeentry.TimeEntry{
				Type:   "linear",
				TaskID: "123",
			},
			expect: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &LinearResolver{}
			res.Init(&config.Config{Linear: tt.cfg})
			desc, err := res.FetchDescription(ctx, tt.entry)
			assert.NoError(t, err)
			assert.Equal(t, tt.expect, desc)
		})
	}
}
