package types_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sumnerevans/tracktime/internal/types"
)

func TestParseTime(t *testing.T) {
	testCases := []struct {
		input    string
		expected *types.Time
	}{
		{"1:00", types.TimeFromMinutes(60)},
		{"1:20", types.TimeFromMinutes(80)},
		{"13:20", types.TimeFromMinutes(800)},
		{"0100", types.TimeFromMinutes(60)},
		{"0120", types.TimeFromMinutes(80)},
		{"1320", types.TimeFromMinutes(800)},
	}
	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			var time types.Time
			err := time.UnmarshalText([]byte(testCase.input))
			assert.NoError(t, err)
			assert.Equal(t, *testCase.expected, time)
		})
	}
}
