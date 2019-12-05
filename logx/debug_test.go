package logx

import (
	"bytes"
	"log"
	"testing"
)

func TestSetup(t *testing.T) {
	tests := []struct {
		name     string
		enable   string
		msg      string
		expected string
		wantErr  bool
	}{
		{
			name:     "success-setup-enable",
			msg:      "this is a test message",
			expected: "DEBUG: this is a test message\n",
			enable:   "true",
		},
		{
			name:   "success-setup-disabled",
			enable: "false",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			log.SetFlags(0)
			log.SetOutput(buf)

			LogxDebug.Set(tt.enable)

			Debug.Print(tt.msg)
			got := string(buf.Bytes())
			if got != tt.expected {
				t.Errorf("Setup did not verify; got %q, want %q", got, tt.expected)
			}
		})
	}
}
