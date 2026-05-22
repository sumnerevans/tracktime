package main

import (
	"reflect"
	"testing"
)

func TestRewriteResumeArgs(t *testing.T) {
	tests := []struct {
		in   []string
		want []string
	}{
		{
			in:   []string{"tt", "resume", "-1"},
			want: []string{"tt", "resume", "--entry=-1"},
		},
		{
			in:   []string{"tt", "resume", "-2"},
			want: []string{"tt", "resume", "--entry=-2"},
		},
		{
			in:   []string{"tt", "resume", "-s", "0930"},
			want: []string{"tt", "resume", "-s", "0930"},
		},
		{
			in:   []string{"tt", "resume", "-s=0930"},
			want: []string{"tt", "resume", "-s=0930"},
		},
		{
			in:   []string{"tt", "resume", "--start", "0930"},
			want: []string{"tt", "resume", "--start", "0930"},
		},
		{
			in:   []string{"tt", "resume", "-n", "2", "-s", "0930"},
			want: []string{"tt", "resume", "-n", "2", "-s", "0930"},
		},
		{
			in:   []string{"tt", "resume", "-s", "0930", "-1"},
			want: []string{"tt", "resume", "-s", "0930", "--entry=-1"},
		},
		{
			in:   []string{"tt", "start", "-1"},
			want: []string{"tt", "start", "-1"},
		},
	}
	for _, tt := range tests {
		got := rewriteResumeArgs(tt.in)
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("rewriteResumeArgs(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}
