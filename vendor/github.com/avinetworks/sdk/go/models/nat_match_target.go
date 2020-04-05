package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NatMatchTarget nat match target
// swagger:model NatMatchTarget
type NatMatchTarget struct {

	// Destination IP of the packet. Field introduced in 18.2.3.
	DestinationIP *IPAddrMatch `json:"destination_ip,omitempty"`

	// Services like port-matching and protocol. Field introduced in 18.2.5.
	Services *ServiceMatch `json:"services,omitempty"`

	// Source IP of the packet. Field introduced in 18.2.3.
	SourceIP *IPAddrMatch `json:"source_ip,omitempty"`
}
