package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServiceMatch service match
// swagger:model ServiceMatch
type ServiceMatch struct {

	// Destination Port of the packet. Field introduced in 18.2.5.
	DestinationPort *PortMatch `json:"destination_port,omitempty"`

	// Source Port of the packet. Field introduced in 18.2.5.
	SourcePort *PortMatch `json:"source_port,omitempty"`
}
