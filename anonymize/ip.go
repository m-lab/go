// Package anonymize provides IP address anonymization tools.
package anonymize

import (
	"flag"
	"fmt"
	"log"
	"net"
)

var (
	// Netblock causes IPv4 addresses to be anonymized up to the
	// /24 level and IPv6 addresses to the /64 level.
	Netblock = Method("netblock")

	// None performs no anonymization. By creating an anonymizer that performs
	// no anonymization, we make it possible to always have the anonymizer code
	// path be used, whether anonymization is actually needed or not, preventing
	// the creation of hundreds of needless `if shouldAnonymize {...}` code
	// blocks.
	None = Method("none")

	// IPAnonymizationFlag is a flag that determines whether IP anonymization is
	// on or off. Its value should be fixed for the duration of a program. This
	// library is not guaranteeed to work properly if you keep switching back
	// and forth between different anonymization schemes. The default is no
	// anonymization.
	IPAnonymizationFlag = None

	// IgnoredIPs is a set of IPs that should be ignored and not anonymized. By
	// default it is the set of local IP addresses. This set should be small, so
	// it is represented as an array because net.IP objects can't be used as map
	// keys.
	//
	// If runtime profiling indicates this causes too much of a slowdown in
	// practice, then the other design that is "near at hand" with Go primitives
	// is to use map[string]struct{} where the string is the IP converted to a
	// string. If that design also causes too much of a slowdown in practice,
	// then we will have to fall back on our algorithms knowledge and build a
	// clever data structure. The 3 main options are:
	//  1. a bloom filter in front of a set object,
	//  2. a 256-ary tree using each successive byte of the IP for each
	//     successive level, or
	//  3. a map[int32]struct{} for v4 addresses and a
	//     map[int64](map[int64]struct{}) for v6 addresses.
	// Likely the last option will be the fastest; ceteris paribus, native ints
	// and language builtins tend to have the best performance.
	//
	// Big-O analysis doesn't buy us much here. There are a small number of
	// local IPs and each one is of length 4 or 16. This means that taking the
	// limit as N goes to infinity doesn't tell us much, because all of the
	// quantities we might call "N" are 16 or less in all realistic scenarios,
	// and 16 is a poor approximation of infinity. Therefore, any claims about
	// relative speed of implementation need to be backed up by experimental
	// evidence.
	IgnoredIPs = []net.IP{}

	// An injected log.Fatal to aid in testing.
	logFatalf = log.Fatalf
)

// Method is an enum suitable for using as a command-line flag. It
// allows only a finite set of values. We can imagine future anonymization
// techniques based on k-anonymity or that completely blot out the IP. We leave
// room for those implementations here, but do not (yet) implement them.
type Method string

// Get is required for all flag.Flag values.
func (m Method) Get() interface{} {
	return m
}

// Set is required for all flag.Flag values.
func (m *Method) Set(s string) error {
	switch Method(s) {
	case Netblock:
		*m = Netblock
	case None:
		*m = None
	default:
		return fmt.Errorf("Uknown anonymization method: %q", s)
	}
	return nil
}

// String is required for all flag.Flag values.
func (m Method) String() string {
	return string(m)
}

func init() {
	flag.Var(&IPAnonymizationFlag, "anonymize.ip", "Valid values are \"none\" and \"netblock\".")

	// Set up the local IP addresses to be ignored by the anonymization system.
	// We want to anonymize our users but not ourselves.
	localAddrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range localAddrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				IgnoredIPs = append(IgnoredIPs, ipnet.IP)
			}
		}
	}
}

// IPAnonymizer is the generic interface for all systems that try and ensure IP
// addresses are not human identifiers. It is a problem with many potential
// subtleties, so we permit multiple implementations. We anonymize the address
// in-place. If you don't want the address to be modified, then make a copy
// before you pass it in.
type IPAnonymizer interface {
	IP(ip net.IP)
	Contains(dst, ip net.IP) bool
}

// New is an IP anonymization factory function that expects you to pass in
// anonymize.IPAnonymizationFlag, which contains the contents of the
// `--anonymize.ip` command-line flag.
//
// If the anonymization method is set to "netblock", then IPv4 addresses will be
// anonymized up to the /24 level and IPv6 addresses to the /64 level. If it is
// set to "none" then no anonymization will be performed. We can imagine future
// anonymization techniques based on k-anonymity or that completely blot out the
// IP. We leave room for those implementations here, but do not (yet) implement
// them.
//
// A program attempting to perform IP anonymization should only ever create one
// IPAnonymizer and use that one anonymizer for all connections. Otherwise, the
// created IPAnonymizer will lack the necessary context to correctly perform
// k-anonymization.
func New(method Method) IPAnonymizer {
	switch method {
	case None:
		return nullIPAnonymizer{}
	case Netblock:
		return netblockAnonymizer{}
	default:
		logFatalf("Unknown anonymization method: %q, exiting to avoid accidentally leaking private data", method)
		panic("This line should only be reached during testing.")
	}
}

// nullIPAnonymizer does nothing.
type nullIPAnonymizer struct{}

func (nullIPAnonymizer) IP(ip net.IP) {}
func (nullIPAnonymizer) Contains(dst, ip net.IP) bool {
	return dst.Equal(ip)
}

// netblockIPAnonymizer restricts v4 addresses to a /24 and v6 addresses to a /64
type netblockAnonymizer struct{}

func (netblockAnonymizer) IP(ip net.IP) {
	if ip == nil {
		return
	}
	for i := range IgnoredIPs {
		if IgnoredIPs[i].Equal(ip) {
			return
		}
	}
	if ip.To4() != nil {
		// Zero out the last byte.  That's ip[3] in the 4-byte v4 representation and ip[15] in the v4-in-v6 representation.
		ip[len(ip)-1] = 0
		return
	}
	if ip.To16() != nil {
		// Truncate IPv6 addresses to the containing /48
		for i := 6; i < 16; i++ {
			ip[i] = 0
		}
		return
	}
	log.Println("The passed in IP address was neither a v4 nor a v6 address:", ip)
	return
}

// Contains determines whether the dst IP (as if netblock anonymized), contains the given dst address.
func (n netblockAnonymizer) Contains(dst, ip net.IP) bool {
	if dst == nil || ip == nil {
		return false
	}
	for i := range IgnoredIPs {
		if IgnoredIPs[i].Equal(dst) {
			return false
		}
	}
	if dst.To4() != nil {
		nn := &net.IPNet{
			IP:   dst,
			Mask: net.CIDRMask(24, 32),
		}
		return nn.Contains(ip)
	}
	if dst.To16() != nil {
		nn := &net.IPNet{
			IP:   dst,
			Mask: net.CIDRMask(48, 128),
		}
		return nn.Contains(ip)
	}
	return false
}
