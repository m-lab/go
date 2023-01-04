package anonymize_test

import (
	"log"
	"net"
	"testing"

	"github.com/go-test/deep"
	"github.com/m-lab/go/anonymize"
	"github.com/m-lab/go/rtx"
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
	verifyNoAnonymization(anonymize.New(anonymize.None), t)
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
	anonymize.New(anonymize.Method("bad_anon_method"))
}

func TestNetblockAnon(t *testing.T) {
	anon := anonymize.New(anonymize.Netblock)

	anon.IP(nil)                  // No crash = success
	anon.IP(net.IP([]byte{1, 2})) // No crash = success

	anonymize.IgnoredIPs = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("1::2")}

	tests := []struct {
		ip   string
		want string
	}{
		{"127.0.0.1", "127.0.0.1"}, // IgnoredIPs should be ignored.
		{"1:0::2", "1::2"},         // IgnoredIPs should be ignored.
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

func TestMethodFlagMethods(t *testing.T) {
	var m anonymize.Method
	rtx.Must(m.Set("none"), "Could not set to none")
	if diff := deep.Equal(anonymize.None, m.Get()); diff != nil {
		t.Error(diff)
	}
	if diff := deep.Equal("none", m.String()); diff != nil {
		t.Error(diff)
	}
	rtx.Must(m.Set("netblock"), "Could not set to netblock")
	if m.Set("badmethod") == nil {
		t.Error("Should have had an error")
	}
	if diff := deep.Equal(anonymize.Netblock, m.Get()); diff != nil {
		t.Error(diff)
	}
	if diff := deep.Equal("netblock", m.String()); diff != nil {
		t.Error(diff)
	}
	log.Println(m)
}

func Example() {
	ip := net.ParseIP("10.10.4.3")
	anon := anonymize.New(anonymize.IPAnonymizationFlag)
	anon.IP(ip)
	log.Println(ip) // Should be "10.10.4.0" if the --anonymize.ip=netblock command-line flag was passed.
}

func TestAnonymizerContains(t *testing.T) {
	anonymize.IgnoredIPs = []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("1::2")}
	tests := []struct {
		name string
		n    anonymize.IPAnonymizer
		dst  net.IP
		ip   net.IP
		want bool
	}{
		// Does the destination IP (as a netblock) contain the given ip?
		{
			name: "success-ipv4",
			n:    anonymize.New(anonymize.Netblock),
			dst:  net.ParseIP("192.168.0.1"),
			ip:   net.ParseIP("192.168.0.2"),
			want: true,
		},
		{
			name: "success-ipv4",
			n:    anonymize.New(anonymize.Netblock),
			dst:  net.ParseIP("192.168.0.2"),
			ip:   net.ParseIP("192.168.0.1"),
			want: true,
		},
		{
			name: "nullanonymizer-ipv4-no-match",
			n:    anonymize.New(anonymize.None),
			dst:  net.ParseIP("192.168.0.1"),
			ip:   net.ParseIP("192.168.0.2"),
			want: false,
		},
		{
			name: "nullanonymizer-ipv4-match",
			n:    anonymize.New(anonymize.None),
			dst:  net.ParseIP("192.168.0.2"),
			ip:   net.ParseIP("192.168.0.2"),
			want: true,
		},
		{
			name: "success-ipv6",
			n:    anonymize.New(anonymize.Netblock),
			dst:  net.ParseIP("fd12:3456:789a:1::1"),
			ip:   net.ParseIP("fd12:3456:789a:1::2"),
			want: true,
		},
		{
			name: "success-ipv6",
			n:    anonymize.New(anonymize.Netblock),
			dst:  net.ParseIP("fd12:3456:789a:1::2"),
			ip:   net.ParseIP("fd12:3456:789a:1::1"),
			want: true,
		},
		{
			name: "success-ignored",
			n:    anonymize.New(anonymize.Netblock),
			dst:  net.ParseIP("127.0.0.1"),
			ip:   net.ParseIP("127.0.0.1"),
			want: false,
		},
		{
			name: "success-nil-dst-arg",
			n:    anonymize.New(anonymize.Netblock),
			dst:  nil,
			ip:   net.ParseIP("127.0.0.1"),
			want: false,
		},
		{
			name: "success-nil-ip-arg",
			n:    anonymize.New(anonymize.Netblock),
			dst:  net.ParseIP("127.0.0.1"),
			ip:   nil,
			want: false,
		},
		{
			name: "error-invalid-byte-array-1-byte-ip",
			n:    anonymize.New(anonymize.Netblock),
			dst:  []byte{'0'},
			ip:   net.ParseIP("127.0.0.1"),
			want: false,
		},
		{
			name: "error-invalid-byte-array-5-byte-ip",
			n:    anonymize.New(anonymize.Netblock),
			dst:  append(net.ParseIP("127.0.0.1"), '0'),
			ip:   net.ParseIP("127.0.0.1"),
			want: false,
		},
		{
			name: "error-invalid-byte-array-17-byte-ip",
			n:    anonymize.New(anonymize.Netblock),
			dst:  append(net.ParseIP("fd12:3456:789a:1::1"), '0'),
			ip:   net.ParseIP("2::1"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.Contains(tt.dst, tt.ip); got != tt.want {
				t.Errorf("netblockAnonymizer.Contains() = %v, want %v", got, tt.want)
			}
		})
	}
}
