// Package host parses v1 and v2 hostnames into their constituent parts. It
// is intended to help in the transition from v1 to v2 names on the platform.
// M-Lab go programs that need to parse hostnames should use this package.
package host

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
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
	var err error

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
	// it from v1 names with four fields.
	switch {
	case len(fields) == 5:
		parts, err = parseHostV3(fields)
		if err != nil {
			return parts, err
		}
	case len(fields) == 4 && len(fields[0]) > 6:
		parts, err = parseHostV2(fields)
		if err != nil {
			return parts, err
		}
	default:
		parts, err = parseHostV1(name)
		if err != nil {
			return parts, err
		}
	}

	return parts, nil
}

// v1 - Example hostnames with field counts when split by '.':
//
//	mlab1.lga01.measurement-lab.org - 4
//	ndt-iupui-mlab1-lga01.measurement-lab.org  - 3
//	ndt.iupui.mlab1.lga01.measurement-lab.org  - 6
func parseHostV1(h string) (Name, error) {
	reV1 := regexp.MustCompile(`(?:[a-z-.]+)?(mlab[1-4]d?)[-.]([a-z]{3}[0-9tc]{2})\.(measurement-lab.org)$`)
	mV1 := reV1.FindAllStringSubmatch(h, -1)
	if len(mV1) != 1 || len(mV1[0]) != 4 {
		return Name{}, fmt.Errorf("invalid v1 hostname: %s", h)
	}
	parts := Name{
		Machine: mV1[0][1],
		Site:    mV1[0][2],
		Project: "",
		Domain:  mV1[0][3],
		Version: "v1",
	}
	return parts, nil
}

// v2 - Example hostnames with field counts when split by '.':
//
//	mlab1-lga01.mlab-oti.measurement-lab.org - 4
//	mlab1-lga01.mlab-oti.measurement-lab.org-d9h6 - 4 (A MIG instance with a random suffix)
//	ndt-mlab1-lga01.mlab-oti.measurement-lab.org-d9h6 - 4 (A MIG instance with a service and random suffix)
//	ndt-iupui-mlab1-lga01.mlab-oti.measurement-lab.org - 4
//	ndt-mlab1-lga01.mlab-oti.measurement-lab.org - 4
func parseHostV2(f []string) (Name, error) {
	if len(f) != 4 || len(f[0]) < 7 {
		return Name{}, fmt.Errorf("invalid v2 hostname: %#v", f)
	}
	sms := strings.Split(f[0], "-")
	var service string
	var machine string
	var site string
	switch len(sms) {
	case 2:
		machine = sms[0]
		site = sms[1]
	case 3:
		service = sms[0]
		machine = sms[1]
		site = sms[2]
	default:
		return Name{}, fmt.Errorf("invalid v2 hostname: %#v", f)
	}
	if !((len(machine) == 5 && unicode.IsDigit(rune(machine[4]))) || (len(machine) == 6 && machine[5] == 'd')) {
		return Name{}, fmt.Errorf("invalid v2 machine name: %#v", f)
	}
	// Fourth site character is always a digit, the fifth is either digit or 't'.
	if len(site) != 5 || !unicode.IsDigit(rune(site[3])) || !(unicode.IsDigit(rune(site[4])) || site[4] == 't') {
		return Name{}, fmt.Errorf("invalid v2 machine name: %#v", f)
	}
	sd := strings.Split(f[3], "-")
	var domain string
	var suffix string
	if len(sd) != 1 && len(sd) != 2 {
		return Name{}, fmt.Errorf("invalid v2 hostname: %#v", f)
	}
	domain = f[2] + "." + sd[0]
	if len(sd) == 2 {
		suffix = "-" + sd[1]
	}
	if domain != "measurement-lab.org" {
		return Name{}, fmt.Errorf("invalid domain: %#v", f)
	}
	parts := Name{
		Service: service,
		Machine: machine,
		Site:    site,
		Project: f[1],
		Domain:  domain,
		Suffix:  suffix,
		Version: "v2",
	}
	return parts, nil
}

// The v3 naming convention is defined in:
// * https://docs.google.com/document/d/1XHgpX7Tbjy_c71TKsFUxb1_ax2624PB4SE9R15OoD_o/edit?#heading=h.s5vpfclyu15x
// The structure follows the pattern:
// * <service>-<IATA><ASN>-<machine>.<organization>.<project>.measurement-lab.org
// * the same rules apply for service, iata, and project names as earlier versions.
// * most ASNs are 16bit numbers, but since 2007 they can be 32bit numbers, allowing up to 10 decimal digits.
// * machine names are 8 character hex encoded IPv4 addresses.
// * site name precedes machine name for readability.
//
// v3 - Example hostnames with field counts when split by '.':
//
//	lga3356-c89ffeef.rnp.autojoin.measurement-lab.org - 5
//	ndt-lga3356-c0a80001.rnp.autojoin.measurement-lab.org - 5
//	ndt-lga3356-040e9f4b.mlab.sandbox.measurement-lab.org - 5
func parseHostV3(f []string) (Name, error) {
	if len(f) != 5 {
		return Name{}, fmt.Errorf("invalid v3 hostname: %#v", f)
	}
	ssm := strings.Split(f[0], "-")
	var service string
	var machine string
	var site string
	switch len(ssm) {
	case 2:
		site = ssm[0]
		machine = ssm[1]
	case 3:
		service = ssm[0]
		site = ssm[1]
		machine = ssm[2]
		if service == "" {
			// There were three fields, but the service field was empty.
			return Name{}, fmt.Errorf("invalid v3 service: %#v", f)
		}
	default:
		return Name{}, fmt.Errorf("invalid v3 hostname: %#v", f)
	}
	if len(machine) != 8 {
		return Name{}, fmt.Errorf("invalid v3 machine: %s", machine)
	}
	if len(site) < 4 || len(site) > 13 {
		return Name{}, fmt.Errorf("invalid v3 site: %s", site)
	}
	for i := range site {
		if i < 3 && !unicode.IsLetter(rune(site[i])) {
			return Name{}, fmt.Errorf("invalid v3 site: %s", site)
		}
		if i >= 3 && !unicode.IsDigit(rune(site[i])) {
			return Name{}, fmt.Errorf("invalid v3 site: %s", site)
		}
	}
	parts := Name{
		Service: service,
		Site:    site,
		Machine: machine,
		Org:     f[1],
		Project: f[2],
		Domain:  f[3] + "." + f[4],
		Version: "v3",
	}

	return parts, nil
}

/*
	reV1 := regexp.MustCompile(`(?:[a-z-.]+)?(mlab[1-4]d?)[-.]([a-z]{3}[0-9tc]{2})\.(measurement-lab.org)$`)
	reV2 := regexp.MustCompile(`([a-z0-9]+)?-?(mlab[1-4]d?)-([a-z]{3}[0-9tc]{2})\.(.*?)\.(measurement-lab.org)(-[a-z0-9]{4})?$`)
	reV3 := regexp.MustCompile(`^(?:([a-z0-9]+)-)?([a-z]{3}[0-9]{1,10})-([a-fA-F0-9]{8})\.(.*?)\.(.*?)\.(measurement-lab.org)$`)
*/

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
