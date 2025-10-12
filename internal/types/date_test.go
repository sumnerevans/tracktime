package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sumnerevans/tracktime/internal/types"
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
			var date types.Date
			err := date.UnmarshalText([]byte(tc.input))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedYear, date.Year(), "year mismatch")
			assert.Equal(t, tc.expectedMonth, date.Month(), "month mismatch")
			assert.Equal(t, tc.expectedDay, date.Day(), "day mismatch")
		})
	}
}

func TestDateUnmarshalTextWeekdays(t *testing.T) {
	// Test all seven weekdays
	weekdays := []struct {
		input   string
		weekday time.Weekday
	}{
		{"sunday", time.Sunday},
		{"monday", time.Monday},
		{"tuesday", time.Tuesday},
		{"wednesday", time.Wednesday},
		{"thursday", time.Thursday},
		{"friday", time.Friday},
		{"saturday", time.Saturday},
	}

	for _, tc := range weekdays {
		t.Run(tc.input, func(t *testing.T) {
			now := time.Now()

			var date types.Date
			err := date.UnmarshalText([]byte(tc.input))
			assert.NoError(t, err)

			// The result should be the target weekday
			assert.Equal(t, tc.weekday, date.Weekday(), "weekday mismatch")

			// If today is the target weekday, result should be today's date
			if now.Weekday() == tc.weekday {
				assert.Equal(t, now.Year(), date.Year(), "should be today's year")
				assert.Equal(t, now.Month(), date.Month(), "should be today's month")
				assert.Equal(t, now.Day(), date.Day(), "should be today's day")
			} else {
				// Otherwise, should be in the past (within last 6 days)
				// Use date-only comparison to avoid time-of-day issues
				nowDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
				dateDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.Local)
				daysDiff := int(nowDate.Sub(dateDate).Hours() / 24)
				assert.True(t, daysDiff >= 1 && daysDiff <= 6,
					"should be 1-6 days in the past, got %d", daysDiff)
			}
		})
	}
}

func TestDateUnmarshalTextWeekdaysCaseInsensitive(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		weekday time.Weekday
	}{
		{"Monday uppercase", "MONDAY", time.Monday},
		{"Tuesday mixed case", "TuEsDaY", time.Tuesday},
		{"Wednesday lowercase", "wednesday", time.Wednesday},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var date types.Date
			err := date.UnmarshalText([]byte(tc.input))
			assert.NoError(t, err)
			assert.Equal(t, tc.weekday, date.Weekday(), "weekday mismatch")
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
			var date types.Date
			err := date.UnmarshalText([]byte(tc.input))
			assert.Error(t, err)
		})
	}
}

func TestDateAddDays(t *testing.T) {
	date := types.NewDate(2024, 3, 15)

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
	date := types.NewDate(2024, 3, 15)

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
