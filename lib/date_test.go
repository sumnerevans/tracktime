package lib_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sumnerevans/tracktime/lib"
)

func TestDateUnmarshalText(t *testing.T) {
	now := time.Now()

	testCases := []struct {
		name          string
		input         string
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		// Special keywords
		{
			name:          "today",
			input:         "today",
			expectedYear:  now.Year(),
			expectedMonth: now.Month(),
			expectedDay:   now.Day(),
		},
		{
			name:          "yesterday",
			input:         "yesterday",
			expectedYear:  now.AddDate(0, 0, -1).Year(),
			expectedMonth: now.AddDate(0, 0, -1).Month(),
			expectedDay:   now.AddDate(0, 0, -1).Day(),
		},

		// Full dates with 4-digit year
		{
			name:          "YYYY-MM-DD format",
			input:         "2024-03-15",
			expectedYear:  2024,
			expectedMonth: time.March,
			expectedDay:   15,
		},
		{
			name:          "YYYY/MM/DD format",
			input:         "2024/03/15",
			expectedYear:  2024,
			expectedMonth: time.March,
			expectedDay:   15,
		},

		// Full dates with 2-digit year
		{
			name:          "YY-MM-DD format",
			input:         "24-03-15",
			expectedYear:  2024,
			expectedMonth: time.March,
			expectedDay:   15,
		},
		{
			name:          "YY/MM/DD format",
			input:         "24/03/15",
			expectedYear:  2024,
			expectedMonth: time.March,
			expectedDay:   15,
		},

		// Month-day format (defaults to current year)
		{
			name:          "MM-DD format",
			input:         "03-15",
			expectedYear:  now.Year(),
			expectedMonth: time.March,
			expectedDay:   15,
		},
		{
			name:          "M-D format (single digit)",
			input:         "3-5",
			expectedYear:  now.Year(),
			expectedMonth: time.March,
			expectedDay:   5,
		},

		// Day only format (defaults to current year and month)
		{
			name:          "DD format",
			input:         "15",
			expectedYear:  now.Year(),
			expectedMonth: now.Month(),
			expectedDay:   15,
		},
		{
			name:          "D format (single digit)",
			input:         "5",
			expectedYear:  now.Year(),
			expectedMonth: now.Month(),
			expectedDay:   5,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var date lib.Date
			err := date.UnmarshalText([]byte(tc.input))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedYear, date.Year(), "year mismatch")
			assert.Equal(t, tc.expectedMonth, date.Month(), "month mismatch")
			assert.Equal(t, tc.expectedDay, date.Day(), "day mismatch")
		})
	}
}

func TestDateUnmarshalTextInvalid(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{"invalid text", "foobar"},
		{"invalid format", "2024-15-40"},
		{"partial invalid", "20-"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var date lib.Date
			err := date.UnmarshalText([]byte(tc.input))
			assert.Error(t, err)
		})
	}
}

func TestDateAddDays(t *testing.T) {
	date := lib.NewDate(2024, 3, 15)

	testCases := []struct {
		name          string
		days          int
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{"add 1 day", 1, 2024, time.March, 16},
		{"add 20 days (crosses month)", 20, 2024, time.April, 4},
		{"subtract 1 day", -1, 2024, time.March, 14},
		{"subtract 15 days (crosses month)", -15, 2024, time.February, 29}, // 2024 is leap year
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := date.AddDays(tc.days)
			assert.Equal(t, tc.expectedYear, result.Year())
			assert.Equal(t, tc.expectedMonth, result.Month())
			assert.Equal(t, tc.expectedDay, result.Day())
		})
	}
}

func TestDateAddMonths(t *testing.T) {
	date := lib.NewDate(2024, 3, 15)

	testCases := []struct {
		name          string
		months        int
		expectedYear  int
		expectedMonth time.Month
		expectedDay   int
	}{
		{"add 1 month", 1, 2024, time.April, 15},
		{"add 10 months (crosses year)", 10, 2025, time.January, 15},
		{"subtract 1 month", -1, 2024, time.February, 15},
		{"subtract 4 months (crosses year)", -4, 2023, time.November, 15},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := date.AddMonths(tc.months)
			assert.Equal(t, tc.expectedYear, result.Year())
			assert.Equal(t, tc.expectedMonth, result.Month())
			assert.Equal(t, tc.expectedDay, result.Day())
		})
	}
}
