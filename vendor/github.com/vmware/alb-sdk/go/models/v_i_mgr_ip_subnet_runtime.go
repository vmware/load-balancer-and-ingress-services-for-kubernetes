package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrIPSubnetRuntime v i mgr IP subnet runtime
// swagger:model VIMgrIPSubnetRuntime
type VIMgrIPSubnetRuntime struct {

	// If true, capable of floating/elastic IP association.
	FipAvailable *bool `json:"fip_available,omitempty"`

	// If fip_available is True, this is list of supported FIP subnets, possibly empty if Cloud does not support such a network list.
	FipSubnetUuids []string `json:"fip_subnet_uuids,omitempty"`

	// If fip_available is True, the list of associated FloatingIP subnets, possibly empty if unsupported or implictly defined by the Cloud. Field introduced in 17.2.1.
	FloatingipSubnets []*FloatingIPSubnet `json:"floatingip_subnets,omitempty"`

	// ip_subnet of VIMgrIPSubnetRuntime.
	IPSubnet *string `json:"ip_subnet,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Placeholder for description of property prefix of obj type VIMgrIPSubnetRuntime field type str  type object
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`

	// True if prefix is primary IP on interface, else false.
	Primary *bool `json:"primary,omitempty"`

	// Number of ref_count.
	RefCount *int32 `json:"ref_count,omitempty"`

	// Number of se_ref_count.
	SeRefCount *int32 `json:"se_ref_count,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
