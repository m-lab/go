package anonymize

import (
	"flag"
	"log"
	"net"
)

var (
	// IPAnonymization is a flag that determines whether IP anonymization is on or
	// off. Its value should be fixed for the duration of a program. This library
	// is not guaranteeed to work properly if you keep switching back and forth
	// between different anonymization schemes.
	IPAnonymization = flag.String("anonymize.ip", "none", "Valid values are \"none\" and \"netblock\".")
)

// IPAnonymizer is the generic interface for all systems that try and ensure IP
// addresses are not human identifiers. It is a problem with many potential
// subtleties, so we permit multiple implementations. We anonymize the address
// in-place. If you don't want the address to be modified, then make a copy
// before you pass it in.
type IPAnonymizer interface {
	IP(ip net.IP)
}

// New is an IP anonymization factory function that respects the
// `--anonymize.ip` command-line flag. If the flag is set to false, then it will
// return the null anonymizer, which actually performs no anonymization at all.
// Through this technique, we make it possible to always have the anonymizer
// code path be used, whether anonymization is actually needed or not,
// preventing the creation of hundreds of needless `if shouldAnonymize {...}`
// code blocks.
//
// If anonymization is turned on, then IPv4 addresses will be anonymized up to
// the /24 level and IPv6 addresses to the /64 level.
func New() IPAnonymizer {
	switch *IPAnonymization {
	case "none":
		return nullIPAnonymizer{}
	case "netblock":
		return netblockAnonymizer{}
	default:
		log.Printf("Unknown anonymization method: %q, using \"none\".\n", *IPAnonymization)
		return nullIPAnonymizer{}
	}
}

// nullIPAnonymizer does nothing.
type nullIPAnonymizer struct{}

func (nullIPAnonymizer) IP(ip net.IP) {}

// netblockIPAnonymizer restricts v4 addresses to a /24 and v6 addresses to a /64
type netblockAnonymizer struct{}

func (netblockAnonymizer) IP(ip net.IP) {
	if ip == nil {
		return
	}
	if ip.To4() != nil {
		// Zero out the last byte.  That's ip[3] in the 4-byte v4 representation and ip[15] in the v4-in-v6 representation.
		ip[len(ip)-1] = 0
		return
	}
	if ip.To16() != nil {
		// Anonymize IPv6 addresses to the containing /64
		for i := 8; i < 16; i++ {
			ip[i] = 0
		}
		return
	}
	log.Println("The passed in IP address was neither a v4 nor a v6 address:", ip)
	return
}
