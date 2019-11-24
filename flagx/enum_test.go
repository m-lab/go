package flagx

import "testing"

func TestEnum_Set(t *testing.T) {
	tests := []struct {
		name    string
		set     string
		Options []string
		wantErr bool
	}{
		{
			name:    "success",
			set:     "warn",
			Options: []string{"debug", "info", "warn"},
		},
		{
			name:    "error-set-invalid-value",
			set:     "invalid-option",
			Options: []string{"debug", "info", "warn"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Enum{
				Options: tt.Options,
			}
			if err := e.Set(tt.set); (err != nil) != tt.wantErr {
				t.Errorf("Enum.Set() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if tt.set != e.Get() {
				t.Errorf("Enum.Get() got = %q, want %q", e.Get(), tt.set)
			}
			if tt.set != e.String() {
				t.Errorf("Enum.Get() got = %q, want %q", e.String(), tt.set)
			}
		})
	}
}
