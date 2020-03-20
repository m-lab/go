package hostname

import (
	"reflect"
	"testing"
)

func TestHostname(t *testing.T) {
	tests := []struct {
		name    string
		want    Hostname
		wantErr bool
	}{
		{
			name: "valid-v1",
			want: Hostname{
				hostname: "mlab1.lol01.measurement-lab.org",
				machine:  "mlab1",
				site:     "lol01",
				project:  "",
				domain:   "measurement-lab.org",
				version:  "v1",
			},
		},
		{
			name: "valid-v2",
			want: Hostname{
				hostname: "mlab1-lol01.mlab-oti.measurement-lab.org",
				machine:  "mlab1",
				site:     "lol01",
				project:  "mlab-oti",
				domain:   "measurement-lab.org",
				version:  "v2",
			},
		},
		{
			name: "invalid-v1-bad-separator",
			want: Hostname{
				hostname: "mlab1=lol01.measurement-lab.org",
				machine:  "mlab1",
				site:     "lol01",
				project:  "",
				domain:   "measurement-lab.org",
				version:  "v1",
			},
			wantErr: true,
		},
		{
			name: "invalid-v1-too-few-parts",
			want: Hostname{
				hostname: "lol01.measurement-lab.org",
				machine:  "",
				site:     "lol01",
				project:  "",
				domain:   "measurement-lab.org",
				version:  "v1",
			},
			wantErr: true,
		},
		{
			name: "invalid-v2-dotted-host",
			want: Hostname{
				hostname: "mlab1.lol01.mlab-staging.measurement-lab.org",
				machine:  "mlab1",
				site:     "lol01",
				project:  "mlab-staging",
				domain:   "measurement-lab.org",
				version:  "v2",
			},
			wantErr: true,
		},
		{
			name: "invalid-v2-too-many-parts",
			want: Hostname{
				hostname: "mlab1-lol01-rofl.mlab-staging.measurement-lab.org",
				machine:  "mlab1",
				site:     "lol01-rofl",
				project:  "mlab-staging",
				domain:   "measurement-lab.org",
				version:  "v2",
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		result, err := Parse(test.want.hostname)
		// If we wanted an err but didn't get one, or vice versa, then fail.
		if (err != nil) != test.wantErr {
			t.Errorf("hostname.Parse() error %v, wantErr %v", err, test.wantErr)
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
