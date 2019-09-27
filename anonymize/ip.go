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
// subtleties, so we permit multiple implementations.
type IPAnonymizer interface {
	IP(ip net.IP) net.IP
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

type nullIPAnonymizer struct{}

func (nullIPAnonymizer) IP(ip net.IP) net.IP {
	return ip
}

type netblockAnonymizer struct{}

func (netblockAnonymizer) IP(ip net.IP) net.IP {
	if ip == nil {
		return nil
	}
	if ip4 := ip.To4(); ip4 != nil {
		// Only copy the first three octets to anonymize a IPv4 address to its containing /24
		return net.IPv4(ip4[0], ip4[1], ip4[2], 0)
	}
	if ip16 := ip.To16(); ip16 != nil {
		// Anonymize IPv6 addresses to the containing /64
		return net.IP{
			ip16[0], ip16[1], ip16[2], ip16[3],
			ip16[4], ip16[5], ip16[6], ip16[7],
			0, 0, 0, 0,
			0, 0, 0, 0,
		}
	}
	log.Println("The passed in IP address was neither a v4 nor a v6 address:", ip)
	return nil
}
