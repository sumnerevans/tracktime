package types_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sumnerevans/tracktime/internal/types"
)

func TestParseMonth(t *testing.T) {
	currentYear := time.Now().Year()
	testCases := []struct {
		input    string
		expected types.Month
	}{
		{"1", types.NewMonth(currentYear, time.January)},
		{"jan", types.NewMonth(currentYear, time.January)},
		{"Jan", types.NewMonth(currentYear, time.January)},
		{"mAr", types.NewMonth(currentYear, time.March)},
		{"08", types.NewMonth(currentYear, time.August)},
		{"11", types.NewMonth(currentYear, time.November)},
		{"january", types.NewMonth(currentYear, time.January)},
		{"January", types.NewMonth(currentYear, time.January)},
	}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			var month types.Month
			err := month.UnmarshalText([]byte(testCase.input))
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, month)
		})
	}
}
