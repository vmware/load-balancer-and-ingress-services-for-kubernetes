package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PlacementNetwork placement network
// swagger:model PlacementNetwork
type PlacementNetwork struct {

	//  It is a reference to an object of type Network.
	// Required: true
	NetworkRef *string `json:"network_ref"`

	// Placeholder for description of property subnet of obj type PlacementNetwork field type str  type object
	// Required: true
	Subnet *IPAddrPrefix `json:"subnet"`
}
