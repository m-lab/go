// Package host parses v1 and v2 hostnames into their constituent parts. It
// is intended to help in the transition from v1 to v2 names on the platform.
// M-Lab go programs that need to parse hostnames should use this package.
package host

import (
	"fmt"
	"regexp"
	"strings"
)

// Name represents an M-Lab hostname and all of its constituent parts.
type Name struct {
	Service string
	Machine string
	Site    string
	Project string
	Org     string
	Domain  string
	Suffix  string
	Version string
}

// Parse parses an M-Lab hostname and breaks it into its constituent parts.
// Parse also accepts service names and discards the service portion of the name.
func Parse(name string) (Name, error) {
	var parts Name

	reV1 := regexp.MustCompile(`(?:[a-z-.]+)?(mlab[1-4]d?)[-.]([a-z]{3}[0-9tc]{2})\.(measurement-lab.org)$`)
	reV2 := regexp.MustCompile(`([a-z0-9]+)?-?(mlab[1-4]d?)-([a-z]{3}[0-9tc]{2})\.(.*?)\.(measurement-lab.org)(-[a-z0-9]{4})?$`)
	// The v3 naming convention is defined in:
	// * https://docs.google.com/document/d/1XHgpX7Tbjy_c71TKsFUxb1_ax2624PB4SE9R15OoD_o/edit?#heading=h.s5vpfclyu15x
	// The structure follows the pattern:
	// * <service>-<IATA><ASN>-<machine>.<organization>.<project>.measurement-lab.org
	// * the same rules apply for service, iata, and project names as earlier versions.
	// * most ASNs are 16bit numbers, but since 2007 they can be 32bit numbers, allowing up to 10 decimal digits.
	// * machine names are 6 byte base64 encoded IPv4 addresses.
	// * site and machine names are reversed for readability.
	reV3 := regexp.MustCompile(`^(?:([a-z0-9]+)-)?([a-z]{3}[0-9]{1,10})-([a-zA-Z0-9]{6})\.(.*?)\.(.*?)\.(measurement-lab.org)$`)

	// Example hostnames with field counts when split by '.':
	// v1
	//   mlab1.lga01.measurement-lab.org - 4
	//   ndt-iupui-mlab1-lga01.measurement-lab.org  - 3
	//   ndt.iupui.mlab1.lga01.measurement-lab.org  - 6
	// v2
	//   mlab1-lga01.mlab-oti.measurement-lab.org - 4
	//   mlab1-lga01.mlab-oti.measurement-lab.org-d9h6 - 4 (A MIG instance with a random suffix)
	//   ndt-mlab1-lga01.mlab-oti.measurement-lab.org-d9h6 - 4 (A MIG instance with a service and random suffix)
	//   ndt-iupui-mlab1-lga01.mlab-oti.measurement-lab.org - 4
	//   ndt-mlab1-lga01.mlab-oti.measurement-lab.org - 4
	// v3
	//   lga3356-BA6fSw.rnp.autojoin.measurement-lab.org - 5
	//   ndt-lga3356-BA6fSw.rnp.autojoin.measurement-lab.org - 5
	//   ndt-lga3356-BA6fSw.mlab.sandbox.measurement-lab.org - 5

	if name == "third-party" {
		// Unconditionally return a Name for third-party origins.
		return Name{
			Machine: "third",
			Site:    "party",
			Version: "v2",
		}, nil
	}

	fields := strings.Split(name, ".")
	if len(fields) < 3 || len(fields) > 6 {
		return parts, fmt.Errorf("invalid hostname: %s", name)
	}

	// v3 names always have 5 fields.
	// v2 names always have 4 fields. And, the first field will always
	// be longer than a machine name e.g. "mlab1", which distinguishes
	// it from v1 name with four fields.
	switch {
	case len(fields) == 5:
		mV3 := reV3.FindAllStringSubmatch(name, -1)
		if len(mV3) != 1 || len(mV3[0]) != 7 {
			return parts, fmt.Errorf("invalid v3 hostname: %s", name)
		}
		parts = Name{
			Service: mV3[0][1],
			Site:    mV3[0][2],
			Machine: mV3[0][3],
			Org:     mV3[0][4],
			Project: mV3[0][5],
			Domain:  mV3[0][6],
			Version: "v3",
		}
	case len(fields) == 4 && len(fields[0]) > 6:
		mV2 := reV2.FindAllStringSubmatch(name, -1)
		if len(mV2) != 1 || len(mV2[0]) != 7 {
			return parts, fmt.Errorf("invalid v2 hostname: %s", name)
		}
		parts = Name{
			Service: mV2[0][1],
			Machine: mV2[0][2],
			Site:    mV2[0][3],
			Project: mV2[0][4],
			Domain:  mV2[0][5],
			Suffix:  mV2[0][6],
			Version: "v2",
		}
	default:
		mV1 := reV1.FindAllStringSubmatch(name, -1)
		if len(mV1) != 1 || len(mV1[0]) != 4 {
			return parts, fmt.Errorf("invalid v1 hostname: %s", name)
		}
		parts = Name{
			Machine: mV1[0][1],
			Site:    mV1[0][2],
			Project: "",
			Domain:  mV1[0][3],
			Version: "v1",
		}
	}

	return parts, nil
}

// Returns a typical M-Lab machine hostname
// Example: mlab2-abc01.mlab-sandbox.measurement-lab.org
func (n Name) String() string {
	switch n.Version {
	case "v3":
		return fmt.Sprintf("%s-%s.%s.%s.%s", n.Site, n.Machine, n.Org, n.Project, n.Domain)
	case "v2":
		return fmt.Sprintf("%s-%s.%s.%s", n.Machine, n.Site, n.Project, n.Domain)
	default:
		return fmt.Sprintf("%s.%s.%s", n.Machine, n.Site, n.Domain)
	}
}

// Returns an M-lab hostname with any service name preserved
// Example: ndt-mlab1-abc01.mlab-sandbox.measurement-lab.org
func (n Name) StringWithService() string {
	if n.Service != "" {
		return fmt.Sprintf("%s-%s", n.Service, n.String())
	} else {
		return n.String()
	}
}

// Returns an M-lab hostname with any suffix preserved
// Example: mlab1-abc01.mlab-sandbox.measurement-lab.org-gz77
func (n Name) StringWithSuffix() string {
	return fmt.Sprintf("%s%s", n.String(), n.Suffix)
}

// Returns an M-lab hostname with any service and suffix preserved
// Example: ndt-mlab1-abc01.mlab-sandbox.measurement-lab.org-gz77
func (n Name) StringAll() string {
	return n.StringWithService() + n.Suffix
}
