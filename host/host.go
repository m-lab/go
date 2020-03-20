// Package host parses v1 and v2 hostnames into their constituent parts. It
// is intended to help in the transition from v1 to v2 names on the platform.
// M-Lab go programs that need to parse hostnames should use this package.
package host

import (
	"fmt"
	"regexp"
)

// Name represents an M-Lab hostname and all of its constituent parts.
type Name struct {
	Machine string
	Site    string
	Project string
	Domain  string
	Version string
}

// Parse parses an M-Lab hostname and breaks it into its constituent parts.
func Parse(name string) (Name, error) {
	var parts Name

	reInit := regexp.MustCompile(`^mlab[1-4]([.-])`)
	reV1 := regexp.MustCompile(`^(mlab[1-4])\.([a-z]{3}[0-9tc]{2})\.(measurement-lab.org)$`)
	reV2 := regexp.MustCompile(`^(mlab[1-4])-([a-z]{3}[0-9tc]{2})\.(.*?)\.(measurement-lab.org)$`)

	mInit := reInit.FindAllStringSubmatch(name, -1)
	if len(mInit) != 1 || len(mInit[0]) != 2 {
		return parts, fmt.Errorf("Invalid hostname: %s", name)
	}

	switch mInit[0][1] {
	case "-":
		mV2 := reV2.FindAllStringSubmatch(name, -1)
		if len(mV2) != 1 || len(mV2[0]) != 5 {
			return parts, fmt.Errorf("Invalid v2 hostname: %s", name)
		}
		parts = Name{
			Machine: mV2[0][1],
			Site:    mV2[0][2],
			Project: mV2[0][3],
			Domain:  mV2[0][4],
			Version: "v2",
		}
	case ".":
		mV1 := reV1.FindAllStringSubmatch(name, -1)
		if len(mV1) != 1 || len(mV1[0]) != 4 {
			return parts, fmt.Errorf("Invalid v1 hostname: %s", name)
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
