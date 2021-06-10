package siteinfo

// Structs corresponding to entities in siteinfo's output formats.
// (https://github.com/m-lab/siteinfo/tree/master/formats/v1/sites)

// Switch is an entity in /v1/sites/switches.json.
type Switch struct {
	AutoNegotiation string `json:"auto_negotation"`
	FlowControl     string `json:"flow_control"`
	IPv4Prefix      string `json:"ipv4_prefix"`
	RSTP            string `json:"rstp"`
	SwitchMake      string `json:"switch_make"`
	SwitchModel     string `json:"switch_model"`
	UplinkPort      string `json:"uplink_port"`
	UplinkSpeed     string `json:"uplink_speed"`
}

type Machine struct {
	Hostname string `json:"hostname"`
	IPv4     string `json:"ipv4"`
	IPv6     string `json:"ipv6"`
	Project  string `json:"project"`
}
