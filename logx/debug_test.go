package logx

import (
	"bytes"
	"log"
	"testing"
)

func TestLogxDebug(t *testing.T) {
	tests := []struct {
		name     string
		enable   string
		msg      string
		expected string
		want     bool
	}{
		{
			name:     "success-setup-enable",
			msg:      "this is a test message",
			expected: "DEBUG: this is a test message\n",
			enable:   "true",
			want:     true,
		},
		{
			name:   "success-setup-disabled",
			enable: "false",
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			log.SetFlags(0)
			log.SetOutput(buf)

			LogxDebug.Set(tt.enable)
			if LogxDebug.Get() != tt.want {
				t.Errorf("LogxDebug.Get return wrong value; got %t, want %t", LogxDebug.Get(), tt.want)
			}

			Debug.Print(tt.msg)
			got := string(buf.Bytes())
			if got != tt.expected {
				t.Errorf("Setup did not verify; got %q, want %q", got, tt.expected)
			}
		})
	}
}
