package logx

import (
	"testing"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:  "success",
			value: "debug",
		},
		{
			name:  "success-empty-value",
			value: "",
		},
		{
			name:    "error-unsupported-option",
			value:   "not-an-option",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			LogxLevel.Value = tt.value
			if err := Setup(); (err != nil) != tt.wantErr {
				t.Errorf("Setup() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoggers(t *testing.T) {
	debug, info, warn := Loggers()
	if debug != Debug {
		t.Errorf("Loggers() debug = %v, want %v", debug, Debug)
	}
	if info != Info {
		t.Errorf("Loggers() info = %v, want %v", info, Info)
	}
	if warn != Warn {
		t.Errorf("Loggers() warn = %v, want %v", warn, Warn)
	}
}
