package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkSubnetInfo network subnet info
// swagger:model NetworkSubnetInfo
type NetworkSubnetInfo struct {

	// Number of free.
	Free *int32 `json:"free,omitempty"`

	// network_name of NetworkSubnetInfo.
	NetworkName *string `json:"network_name,omitempty"`

	// Unique object identifier of network.
	NetworkUUID *string `json:"network_uuid,omitempty"`

	// Placeholder for description of property subnet of obj type NetworkSubnetInfo field type str  type object
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`

	// Number of total.
	Total *int32 `json:"total,omitempty"`

	// Number of used.
	Used *int32 `json:"used,omitempty"`
}
