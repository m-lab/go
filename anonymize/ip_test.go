package anonymize_test

import (
	"log"
	"net"
	"testing"

	"github.com/m-lab/go/anonymize"
)

func verifyNoAnonymization(doesNoAnon anonymize.IPAnonymizer, t *testing.T) {
	ip := net.ParseIP("10.10.4.3")
	doesNoAnon.IP(ip)
	if ip.String() != "10.10.4.3" {
		t.Errorf("anonymizedIP (%s) should be 10.10.4.3", ip.String())
	}

	doesNoAnon.IP(nil) // No crash = success.
}

func TestNoAnon(t *testing.T) {
	anonymize.SetFlag("none")
	verifyNoAnonymization(anonymize.New(), t)
}

func TestBadAnonName(t *testing.T) {
	calls := 0
	revert := anonymize.SetLogFatalf(func(string, ...interface{}) {
		calls++
	})
	defer revert()
	defer func() {
		r := recover()
		if r == nil {
			t.Error("A bad anonymization method should cause a panic, but it did not.")
		}
		if calls == 0 {
			t.Error("calls should not be zero")
		}
	}()
	anonymize.SetFlag("bad_anon_method")
	anonymize.New()
}

func TestNetblockAnon(t *testing.T) {
	anonymize.SetFlag("netblock")
	anon := anonymize.New()

	anon.IP(nil)                  // No crash = success
	anon.IP(net.IP([]byte{1, 2})) // No crash = success

	tests := []struct {
		ip   string
		want string
	}{
		{"127.0.0.1", "127.0.0.0"},
		{"10.1.2.3", "10.1.2.0"},
		{"255.255.255.255", "255.255.255.0"},
		{"0:1:2:3:4:5:6:7", "0:1:2:3::"},
		{"aaaa:aaab:aaac:aaad:aaae:aaaf:aaa1:aaa1", "aaaa:aaab:aaac:aaad::"},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			ip := net.ParseIP(tt.ip)
			anon.IP(ip)
			if ip.String() != tt.want {
				t.Errorf("netblockAnonymizer.IP() = %q, want %q", ip.String(), tt.want)
			}
		})
	}
}

func Example() {
	ip := net.ParseIP("10.10.4.3")
	anon := anonymize.New()
	anon.IP(ip)
	log.Println(ip) // Should be "10.10.4.0" if the --anonymize.ip=netblock command-line flag was passed.
}
