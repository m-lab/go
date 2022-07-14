// Package timex provides extensions to Go's time package.
package timex_test

import (
	"testing"
	"time"

	"github.com/m-lab/go/timex"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name   string
		format string
		input  time.Time
		want   string
	}{
		{
			name:   "YYYYMMDD",
			format: timex.YYYYMMDD,
			input:  time.Date(2019, time.April, 03, 0, 0, 0, 0, time.UTC),
			want:   "20190403",
		},
		{
			name:   "YYYYMMDDWithSlash",
			format: timex.YYYYMMDDWithSlash,
			input:  time.Date(2019, time.April, 03, 0, 0, 0, 0, time.UTC),
			want:   "2019/04/03",
		},
		{
			name:   "YYYYMMDDWithDash",
			format: timex.YYYYMMDDWithDash,
			input:  time.Date(2019, time.April, 03, 0, 0, 0, 0, time.UTC),
			want:   "2019-04-03",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.input.Format(tt.format)
			if tt.want != got {
				t.Errorf("timex.%s produced wrong format; got %q, want %q", tt.name, got, tt.want)
			}
		})
	}
}
