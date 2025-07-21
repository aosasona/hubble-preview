package lib

import (
	"testing"
	"time"
)

func Test_FormatToHumanReadableDuration(t *testing.T) {
	type test struct {
		input  time.Duration
		output string
	}

	tests := []test{
		{input: 0, output: "0 seconds"},
		{input: time.Second, output: "1 second"},
		{input: time.Minute, output: "1 minute"},
		{input: time.Hour, output: "1 hour"},
		{input: 24 * time.Hour, output: "1 day"},
		{
			input:  2*time.Hour + 3*time.Minute + 4*time.Second,
			output: "2 hours, 3 minutes and 4 seconds",
		},
		{
			input:  2*24*time.Hour + 2*time.Hour + 3*time.Minute + 4*time.Second,
			output: "2 days, 2 hours, 3 minutes and 4 seconds",
		},
		{
			input:  2*24*time.Hour + 3*time.Hour + 3*time.Minute + 4*time.Second,
			output: "2 days, 3 hours, 3 minutes and 4 seconds",
		},
		{
			input:  2*24*time.Hour + 3*time.Hour + 4*time.Minute + 4*time.Second,
			output: "2 days, 3 hours, 4 minutes and 4 seconds",
		},
		{
			input:  7 * 24 * time.Hour,
			output: "7 days",
		},
	}

	for _, tc := range tests {
		t.Run(tc.output, func(t *testing.T) {
			result := ToHumanReadableDuration(tc.input)
			if result != tc.output {
				t.Errorf("expected %s, got %s", tc.output, result)
			}
		})
	}
}
