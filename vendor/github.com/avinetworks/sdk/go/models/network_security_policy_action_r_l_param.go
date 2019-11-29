package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkSecurityPolicyActionRLParam network security policy action r l param
// swagger:model NetworkSecurityPolicyActionRLParam
type NetworkSecurityPolicyActionRLParam struct {

	// Maximum number of connections or requests or packets to be rate limited instantaneously.
	// Required: true
	BurstSize *int32 `json:"burst_size"`

	// Maximum number of connections or requests or packets per second. Allowed values are 1-4294967295.
	// Required: true
	MaxRate *int32 `json:"max_rate"`
}
