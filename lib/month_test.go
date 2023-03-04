package lib_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sumnerevans/tracktime/lib"
)

func TestParseMonth(t *testing.T) {
	currentYear := time.Now().Year()
	testCases := []struct {
		input    string
		expected lib.Month
	}{
		{"1", lib.NewMonth(currentYear, time.January)},
		{"jan", lib.NewMonth(currentYear, time.January)},
		{"Jan", lib.NewMonth(currentYear, time.January)},
		{"mAr", lib.NewMonth(currentYear, time.March)},
		{"08", lib.NewMonth(currentYear, time.August)},
		{"11", lib.NewMonth(currentYear, time.November)},
		{"january", lib.NewMonth(currentYear, time.January)},
		{"January", lib.NewMonth(currentYear, time.January)},
	}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			var month lib.Month
			err := month.UnmarshalText([]byte(testCase.input))
			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, month)
		})
	}
}
