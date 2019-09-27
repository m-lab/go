package anonymize_test

import (
	"log"
	"net"
	"testing"

	"github.com/m-lab/go/anonymize"
)

func verifyPassThrough(anon anonymize.IPAnonymizer, t *testing.T) {
	ip := net.ParseIP("10.10.4.3")
	anon.IP(ip)
	if ip.String() != "10.10.4.3" {
		t.Errorf("anonymizedIP (%s) should be 10.10.4.3", ip.String())
	}

	anon.IP(nil) // No crash = success.
}

func TestNoAnon(t *testing.T) {
	*anonymize.IPAnonymization = "none"
	verifyPassThrough(anonymize.New(), t)
}

func TestBadAnonName(t *testing.T) {
	*anonymize.IPAnonymization = "bad_anon_method"
	verifyPassThrough(anonymize.New(), t)
}

func TestNetblockAnon(t *testing.T) {
	*anonymize.IPAnonymization = "netblock"
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
