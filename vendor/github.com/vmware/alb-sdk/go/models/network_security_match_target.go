package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkSecurityMatchTarget network security match target
// swagger:model NetworkSecurityMatchTarget
type NetworkSecurityMatchTarget struct {

	// Placeholder for description of property client_ip of obj type NetworkSecurityMatchTarget field type str  type object
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// Matches the source port of incoming packets in the client side traffic. Field introduced in 20.1.3.
	ClientPort *PortMatchGeneric `json:"client_port,omitempty"`

	// Matches the geo information of incoming packets in the client side traffic. Field introduced in 21.1.1. Maximum of 1 items allowed.
	GeoMatches []*GeoMatch `json:"geo_matches,omitempty"`

	//  Field introduced in 20.1.1. Allowed in Basic edition, Essentials edition, Enterprise edition.
	IPReputationType *IPReputationTypeMatch `json:"ip_reputation_type,omitempty"`

	// Placeholder for description of property microservice of obj type NetworkSecurityMatchTarget field type str  type object
	Microservice *MicroServiceMatch `json:"microservice,omitempty"`

	// Placeholder for description of property vs_port of obj type NetworkSecurityMatchTarget field type str  type object
	VsPort *PortMatch `json:"vs_port,omitempty"`
}
