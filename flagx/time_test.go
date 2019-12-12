package flagx

import (
	"fmt"
	"testing"
	"time"
)

func TestDateTime_Set(t *testing.T) {
	tests := []struct {
		name       string
		arg        string
		wantTime   time.Time
		wantFormat string
		wantErr    bool
	}{
		{
			name:       "success-date",
			arg:        "2019-03-30",
			wantTime:   time.Date(2019, 3, 30, 0, 0, 0, 0, time.UTC),
			wantFormat: "2019-03-30",
		},
		{
			name:       "success-date-ambiguous",
			arg:        "2019-03-01",
			wantTime:   time.Date(2019, 3, 1, 0, 0, 0, 0, time.UTC),
			wantFormat: "2019-03-01",
		},
		{
			name:       "success-datetime",
			arg:        "2019-03-30T12:34:56",
			wantTime:   time.Date(2019, 3, 30, 12, 34, 56, 0, time.UTC),
			wantFormat: "2019-03-30T12:34:56",
		},
		{
			name:       "success-datetime-seconds",
			arg:        "1553949296001",
			wantTime:   time.Date(2019, 3, 30, 12, 34, 56, 1000000, time.UTC),
			wantFormat: "1553949296001",
		},
		{
			name:    "error-bad-date",
			arg:     "019/06/30",
			wantErr: true,
		},

		{
			name:    "error-bad-datetime-ambiguous",
			arg:     "01/01/19",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &DateTime{}
			if err := f.Set(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("DateTime.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if f.Time != tt.wantTime {
				t.Errorf("DateTime.Set() raw time not equal; got = %s, want %s", f.Time, tt.wantTime)
			}
			if f.String() != tt.wantFormat {
				t.Errorf("DateTime.String() format not equal; got = %q, want %q", f.String(), tt.wantFormat)
			}
			f2 := &DateTime{}
			f2.Set(f.String())
			fmt.Println(f2.String())
		})
	}
}

func TestTime_Set(t *testing.T) {
	tests := []struct {
		name       string
		arg        string
		wantTime   Time
		wantFormat string
		wantErr    bool
	}{
		{
			name:       "success-time",
			arg:        "19:03:30",
			wantTime:   Time{19, 3, 30},
			wantFormat: "19:03:30",
		},
		{
			name:    "error-bad-time",
			arg:     "19.06.30",
			wantErr: true,
		},
		{
			name:    "error-bad-length",
			arg:     "19:06:3",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Time{}
			if err := f.Set(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("Time.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if f != tt.wantTime {
				t.Errorf("Time.Set() parsed time not equal; got = %#v, want %#v", f, tt.wantTime)
			}
			if f.String() != tt.wantFormat {
				t.Errorf("Time.String() format not equal; got = %q, want %q", f.String(), tt.wantFormat)
			}
		})
	}
}
