package host

import (
	"reflect"
	"testing"

	"github.com/m-lab/go/rtx"
)

func TestName(t *testing.T) {
	tests := []struct {
		name     string
		hostname string
		want     Name
		wantErr  bool
	}{
		{
			name:     "valid-v1",
			hostname: "mlab1.lol01.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "",
				Domain:  "measurement-lab.org",
				Version: "v1",
			},
		},
		{
			name:     "valid-v2",
			hostname: "mlab1-lol01.mlab-oti.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-oti",
				Domain:  "measurement-lab.org",
				Version: "v2",
			},
		},
		{
			name:     "valid-v2-with-suffix",
			hostname: "mlab1-lol01.mlab-oti.measurement-lab.org-a9b8",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-oti",
				Domain:  "measurement-lab.org",
				Suffix:  "a9b8",
				Version: "v2",
			},
		},
		{
			name:     "valid-v1-bmc",
			hostname: "mlab1d.lol01.measurement-lab.org",
			want: Name{
				Machine: "mlab1d",
				Site:    "lol01",
				Domain:  "measurement-lab.org",
				Version: "v1",
			},
		},
		{
			name:     "valid-v2-bmc",
			hostname: "mlab1d-lol01.mlab-oti.measurement-lab.org",
			want: Name{
				Machine: "mlab1d",
				Site:    "lol01",
				Project: "mlab-oti",
				Domain:  "measurement-lab.org",
				Version: "v2",
			},
		},
		{
			name:     "valid-v1-with-ndt-flat",
			hostname: "ndt-iupui-mlab1-lol01.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Domain:  "measurement-lab.org",
				Version: "v1",
			},
		},
		{
			name:     "valid-v1-with-ndt-regular",
			hostname: "ndt.iupui.mlab1.lol01.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Domain:  "measurement-lab.org",
				Version: "v1",
			},
		},
		{
			name:     "valid-v2-with-ndt",
			hostname: "ndt-iupui-mlab1-lol01.mlab-oti.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-oti",
				Domain:  "measurement-lab.org",
				Version: "v2",
			},
		},
		{
			name:     "valid-v2-with-ndt-short",
			hostname: "ndt-mlab1-lol01.mlab-oti.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-oti",
				Domain:  "measurement-lab.org",
				Version: "v2",
			},
		},
		{
			name:     "valid-v2-with-ndt-short-with-suffix",
			hostname: "ndt-mlab1-lol01.mlab-oti.measurement-lab.org-q44c",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-oti",
				Domain:  "measurement-lab.org",
				Suffix:  "q44c",
				Version: "v2",
			},
		},
		{
			name:     "invalid-too-few-separators",
			hostname: "mlab1-lol01-measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v2-bad-domain",
			hostname: "mlab1-lol01.mlab-sandbox.measurementlab.net",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v1-bad-separator",
			hostname: "mlab1=lol01.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v1-too-few-parts",
			hostname: "lol01.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v2-dotted-host",
			hostname: "mlab1.lol01.mlab-staging.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v2-too-many-parts",
			hostname: "mlab1-lol01-rofl.mlab-staging.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := Parse(test.hostname)
			// If we wanted an err but didn't get one, or vice versa, then fail.
			if (err != nil) != test.wantErr {
				t.Errorf("host.Parse() error %v, wantErr %v", err, test.wantErr)
			}
			if test.wantErr {
				return
			}
			if !reflect.DeepEqual(result, test.want) {
				t.Errorf("\nUnexpected result. Got:\n%+v\nExpected:\n%+v", result, test.want)
			}
		})
	}
}

func TestName_String(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "mlab1.foo01.measurement-lab.org"},
		{name: "mlab1-foo01.mlab-sandbox.measurement-lab.org"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := Parse(tt.name)
			rtx.Must(err, "Failed to parse: %s", tt.name)
			if got := n.String(); got != tt.name {
				t.Errorf("Name.String() = %v, want %v", got, tt.name)
			}
		})
	}
}
