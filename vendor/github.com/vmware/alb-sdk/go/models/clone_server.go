package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloneServer clone server
// swagger:model CloneServer
type CloneServer struct {

	// IP Address of the Clone Server. Field introduced in 17.1.1.
	IPAddress *IPAddr `json:"ip_address,omitempty"`

	// MAC Address of the Clone Server. Field introduced in 17.1.1.
	Mac *string `json:"mac,omitempty"`

	// Network to clone the traffic to. It is a reference to an object of type Network. Field introduced in 17.1.1.
	NetworkRef *string `json:"network_ref,omitempty"`

	// Subnet of the network to clone the traffic to. Field introduced in 17.1.1.
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`
}
