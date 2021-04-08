package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IptableRule iptable rule
// swagger:model IptableRule
type IptableRule struct {

	//  Enum options - ACCEPT, DROP, REJECT, DNAT, MASQUERADE.
	// Required: true
	Action *string `json:"action"`

	// Placeholder for description of property dnat_ip of obj type IptableRule field type str  type object
	DnatIP *IPAddr `json:"dnat_ip,omitempty"`

	// Placeholder for description of property dst_ip of obj type IptableRule field type str  type object
	DstIP *IPAddrPrefix `json:"dst_ip,omitempty"`

	// Placeholder for description of property dst_port of obj type IptableRule field type str  type object
	DstPort *PortRange `json:"dst_port,omitempty"`

	// input_interface of IptableRule.
	InputInterface *string `json:"input_interface,omitempty"`

	// output_interface of IptableRule.
	OutputInterface *string `json:"output_interface,omitempty"`

	//  Enum options - PROTO_TCP, PROTO_UDP, PROTO_ICMP, PROTO_ALL.
	Proto *string `json:"proto,omitempty"`

	// Placeholder for description of property src_ip of obj type IptableRule field type str  type object
	SrcIP *IPAddrPrefix `json:"src_ip,omitempty"`

	// Placeholder for description of property src_port of obj type IptableRule field type str  type object
	SrcPort *PortRange `json:"src_port,omitempty"`

	// tag of IptableRule.
	Tag *string `json:"tag,omitempty"`
}
