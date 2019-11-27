package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkSecurityMatchTarget network security match target
// swagger:model NetworkSecurityMatchTarget
type NetworkSecurityMatchTarget struct {

	// Placeholder for description of property client_ip of obj type NetworkSecurityMatchTarget field type str  type object
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Placeholder for description of property microservice of obj type NetworkSecurityMatchTarget field type str  type object
	Microservice *MicroServiceMatch `json:"microservice,omitempty"`

	// Placeholder for description of property vs_port of obj type NetworkSecurityMatchTarget field type str  type object
	VsPort *PortMatch `json:"vs_port,omitempty"`
}
