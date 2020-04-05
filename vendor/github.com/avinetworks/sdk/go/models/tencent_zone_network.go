package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TencentZoneNetwork tencent zone network
// swagger:model TencentZoneNetwork
type TencentZoneNetwork struct {

	// Availability zone. Field introduced in 18.2.3.
	// Required: true
	AvailabilityZone *string `json:"availability_zone"`

	// Usable networks for Virtual IP. If VirtualService does not specify a network and auto_allocate_ip is set, then the first available network from this list will be chosen for IP allocation. Field introduced in 18.2.3.
	// Required: true
	UsableSubnetID *string `json:"usable_subnet_id"`
}
