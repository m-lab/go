package anonymize

import (
	"flag"
	"log"
	"net"
)

var (
	// ipAnonymization is a flag that determines whether IP anonymization is on or
	// off. Its value should be fixed for the duration of a program. This library
	// is not guaranteeed to work properly if you keep switching back and forth
	// between different anonymization schemes. It is not exported to make changing
	// its value difficult.
	ipAnonymization = flag.String("anonymize.ip", "none", "Valid values are \"none\" and \"netblock\".")

	logFatalf = log.Fatalf
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
// If the anonymization method is set to "netblock", then IPv4 addresses will be
// anonymized up to the /24 level and IPv6 addresses to the /64 level. If it is
// set to "none" then no anonymization will be performed. We can imagine future
// anonymization techniques based on k-anonymity or that completely blot out the
// IP. We leave room for those implementations here, but do not (yet) implement
// them.
func New() IPAnonymizer {
	switch *ipAnonymization {
	case "none":
		return nullIPAnonymizer{}
	case "netblock":
		return netblockAnonymizer{}
	default:
		logFatalf("Unknown anonymization method: %q, exiting to avoid accidentally leaking private data", *ipAnonymization)
		panic("This line should only be reached during testing.")
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
