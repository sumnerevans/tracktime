package lib_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sumnerevans/tracktime/lib"
)

func TestParseTime(t *testing.T) {
	testCases := []struct {
		input    string
		expected *lib.Time
	}{
		{"1:00", lib.TimeFromMinutes(60)},
		{"1:20", lib.TimeFromMinutes(80)},
		{"13:20", lib.TimeFromMinutes(800)},
		{"0100", lib.TimeFromMinutes(60)},
		{"0120", lib.TimeFromMinutes(80)},
		{"1320", lib.TimeFromMinutes(800)},
	}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			var time lib.Time
			err := time.UnmarshalText([]byte(testCase.input))
			assert.NoError(t, err)
			assert.Equal(t, *testCase.expected, time)
		})
	}
}
