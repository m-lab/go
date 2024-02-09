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
			hostname: "mlab1-lol01.mlab-sandbox.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-sandbox",
				Domain:  "measurement-lab.org",
				Version: "v2",
			},
		},
		{
			name:     "valid-v2-with-suffix",
			hostname: "mlab1-lol01.mlab-sandbox.measurement-lab.org-a9b8",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-sandbox",
				Domain:  "measurement-lab.org",
				Suffix:  "-a9b8",
				Version: "v2",
			},
		},
		{
			name:     "valid-v2-with-service",
			hostname: "ndt-mlab1-lol01.mlab-sandbox.measurement-lab.org",
			want: Name{
				Service: "ndt",
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-sandbox",
				Domain:  "measurement-lab.org",
				Version: "v2",
			},
		},
		{
			name:     "valid-v2-with-service-and-suffix",
			hostname: "ndt-mlab1-lol01.mlab-sandbox.measurement-lab.org-a9b8",
			want: Name{
				Service: "ndt",
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-sandbox",
				Domain:  "measurement-lab.org",
				Suffix:  "-a9b8",
				Version: "v2",
			},
		},
		{
			name:     "invalid-v2-with-extra-suffix",
			hostname: "ndt-mlab1-lol01.mlab-sandbox.measurement-lab.org-a9b8-abcd",
			wantErr:  true,
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
			hostname: "mlab1d-lol01.mlab-sandbox.measurement-lab.org",
			want: Name{
				Machine: "mlab1d",
				Site:    "lol01",
				Project: "mlab-sandbox",
				Domain:  "measurement-lab.org",
				Version: "v2",
			},
		},
		{
			name:     "invalid-v2-no-service-site-fields",
			hostname: "ndtmlab1lol01.mlab-sandbox.measurement-lab.org",
			wantErr:  true,
		},
		{
			name:     "invalid-v2-bad-machine-name",
			hostname: "ndt-mlab12-lol01.mlab-sandbox.measurement-lab.org",
			wantErr:  true,
		},
		{
			name:     "valid-v2-third-party",
			hostname: "third-party",
			want: Name{
				Machine: "third",
				Site:    "party",
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
			name:     "invalid-v1-third-party",
			hostname: "third-party.measurement-lab.org",
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
		{
			name:     "valid-v3-machine",
			hostname: "lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
			want: Name{
				Machine: "abcdef01",
				Site:    "lol12345",
				Org:     "mlab",
				Project: "sandbox",
				Domain:  "measurement-lab.org",
				Version: "v3",
			},
		},
		{
			name:     "valid-v3-service",
			hostname: "ndt-lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
			want: Name{
				Service: "ndt",
				Machine: "abcdef01",
				Site:    "lol12345",
				Org:     "mlab",
				Project: "sandbox",
				Domain:  "measurement-lab.org",
				Version: "v3",
			},
		},
		{
			name:     "invalid-v3-too-long-asn-machine",
			hostname: "lol12345678901-abcdef01.mlab.sandbox.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v3-too-long-asn-service",
			hostname: "ndt-lol12345678901-abcdef01.mlab.sandbox.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v3-site-too-long",
			hostname: "abcd12345-abcdef01.mlab.sandbox.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v3-site-too-short",
			hostname: "ab12345-abcdef01.mlab.sandbox.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v3-missing-service",
			hostname: "-abc12345-abcdef01.mlab.sandbox.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v3-missing-site",
			hostname: "ndt--abcdef01.mlab.sandbox.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v3-missing-machine",
			hostname: "ndt-abc1234-.mlab.sandbox.measurement-lab.org",
			want:     Name{},
			wantErr:  true,
		},
		{
			name:     "invalid-v3-machine-too-long",
			hostname: "abc12345-abcdef789.mlab.sandbox.measurement-lab.org",
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
				t.Errorf("\nUnexpected result. Got:\n%#v\nExpected:\n%#v", result, test.want)
			}
		})
	}
}

func TestName_String(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "mlab1.foo01.measurement-lab.org",
			want: "mlab1.foo01.measurement-lab.org",
		},
		{
			name: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "ndt-lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
			want: "lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := Parse(tt.name)
			rtx.Must(err, "Failed to parse: %s", tt.name)
			if got := n.String(); got != tt.want {
				t.Errorf("Name.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName_StringWithService(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org",
			want: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
			want: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "ndt-lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
			want: "ndt-lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := Parse(tt.name)
			rtx.Must(err, "Failed to parse: %s", tt.name)
			if got := n.StringWithService(); got != tt.want {
				t.Errorf("Name.StringWithService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName_StringWithSuffix(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
		},
		{
			name: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
		},
		{
			name: "ndt-lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
			want: "lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := Parse(tt.name)
			rtx.Must(err, "Failed to parse: %s", tt.name)
			if got := n.StringWithSuffix(); got != tt.want {
				t.Errorf("Name.StringWithSuffix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestName_StringAll(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org",
			want: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org",
		},
		{
			name: "mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
			want: "mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
		},
		{
			name: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
			want: "ndt-mlab1-foo01.mlab-sandbox.measurement-lab.org-qf8y",
		},
		{
			name: "ndt-lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
			want: "ndt-lol12345-abcdef01.mlab.sandbox.measurement-lab.org",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n, err := Parse(tt.name)
			rtx.Must(err, "Failed to parse: %s", tt.name)
			if got := n.StringAll(); got != tt.want {
				t.Errorf("Name.StringAll() = %v, want %v", got, tt.want)
			}
		})
	}
}
