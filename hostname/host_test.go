package host

import (
	"reflect"
	"testing"
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
			name:     "invalid-v1-bad-separator",
			hostname: "mlab1=lol01.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "",
				Domain:  "measurement-lab.org",
				Version: "v1",
			},
			wantErr: true,
		},
		{
			name:     "invalid-v1-too-few-parts",
			hostname: "lol01.measurement-lab.org",
			want: Name{
				Machine: "",
				Site:    "lol01",
				Project: "",
				Domain:  "measurement-lab.org",
				Version: "v1",
			},
			wantErr: true,
		},
		{
			name:     "invalid-v2-dotted-host",
			hostname: "mlab1.lol01.mlab-staging.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01",
				Project: "mlab-staging",
				Domain:  "measurement-lab.org",
				Version: "v2",
			},
			wantErr: true,
		},
		{
			name:     "invalid-v2-too-many-parts",
			hostname: "mlab1-lol01-rofl.mlab-staging.measurement-lab.org",
			want: Name{
				Machine: "mlab1",
				Site:    "lol01-rofl",
				Project: "mlab-staging",
				Domain:  "measurement-lab.org",
				Version: "v2",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		result, err := Parse(test.hostname)
		// If we wanted an err but didn't get one, or vice versa, then fail.
		if (err != nil) != test.wantErr {
			t.Errorf("host.Parse() error %v, wantErr %v", err, test.wantErr)
		}
		// If we wanted an err, go not further since later tests will fail.
		if test.wantErr == true {
			continue
		}
		if !reflect.DeepEqual(result, test.want) {
			t.Errorf("\nUnexpected result. Got:\n%+v\nExpected:\n%+v", result, test.want)
		}

	}
}
