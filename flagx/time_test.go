package flagx

import (
	"testing"
)

func TestDateTime_Set(t *testing.T) {
	tests := []struct {
		name       string
		arg        string
		wantDate   Date
		wantTime   Time
		wantFormat string
		wantErr    bool
	}{
		{
			name:       "success-date",
			arg:        "2019-03-30",
			wantDate:   Date{2019, 3, 30},
			wantFormat: "2019-03-30T00:00:00",
		},
		{
			name:       "success-datetime",
			arg:        "2019-03-30T12:34:56",
			wantDate:   Date{2019, 3, 30},
			wantTime:   Time{12, 34, 56},
			wantFormat: "2019-03-30T12:34:56",
		},
		{
			name:    "error-bad-date",
			arg:     "2019/06/30",
			wantErr: true,
		},

		{
			name:    "error-bad-datetime-separator",
			arg:     "2019-06-30x12:34:56",
			wantErr: true,
		},
		{
			name:    "error-bad-datetime-length",
			arg:     "2019-06-30T12:30",
			wantErr: true,
		},
		{
			name:    "error-bad-time",
			arg:     "2019-06-30T12/30/00",
			wantErr: true,
		},
		{
			name:    "error-bad-date-with-right-length",
			arg:     "2019/06/30T12:30:00",
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
			if f.Date != tt.wantDate {
				t.Errorf("DateTime.Set() parsed time not equal; got = %#v, want %#v", f.Date, tt.wantDate)
			}
			if f.Time != tt.wantTime {
				t.Errorf("DateTime.Set() raw time not equal; got = %#v, want %#v", f.Time, tt.wantTime)
			}
			if f.String() != tt.wantFormat {
				t.Errorf("DateTime.String() format not equal; got = %q, want %q", f.String(), tt.wantFormat)
			}
		})
	}
}

func TestDate_Set(t *testing.T) {
	tests := []struct {
		name       string
		arg        string
		wantDate   Date
		wantFormat string
		wantErr    bool
	}{
		{
			name:       "success-date",
			arg:        "2019-03-30",
			wantDate:   Date{2019, 3, 30},
			wantFormat: "2019-03-30",
		},
		{
			name:    "error-bad-date",
			arg:     "2019/06/30",
			wantErr: true,
		},
		{
			name:    "error-bad-length",
			arg:     "2019-06-3",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := Date{}
			if err := f.Set(tt.arg); (err != nil) != tt.wantErr {
				t.Errorf("Date.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if f != tt.wantDate {
				t.Errorf("Date.Set() parsed time not equal; got = %#v, want %#v", f, tt.wantDate)
			}
			if f.String() != tt.wantFormat {
				t.Errorf("Date.String() format not equal; got = %q, want %q", f.String(), tt.wantFormat)
			}
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
			name:       "success-date",
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
