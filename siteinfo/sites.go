package siteinfo

// This package contains the structs corresponding the siteinfo's v1 output
// formats. (https://github.com/m-lab/siteinfo/tree/master/formats/v1)

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
