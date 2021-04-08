package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AwsZoneNetwork aws zone network
// swagger:model AwsZoneNetwork
type AwsZoneNetwork struct {

	// Availability zone. Field introduced in 17.1.3.
	// Required: true
	AvailabilityZone *string `json:"availability_zone"`

	// Usable networks for Virtual IP. If VirtualService does not specify a network and auto_allocate_ip is set, then the first available network from this list will be chosen for IP allocation. Field introduced in 17.1.3.
	UsableNetworkUuids []string `json:"usable_network_uuids,omitempty"`
}
