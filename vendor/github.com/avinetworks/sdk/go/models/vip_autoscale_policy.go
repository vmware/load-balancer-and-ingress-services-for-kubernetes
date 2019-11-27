package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VipAutoscalePolicy vip autoscale policy
// swagger:model VipAutoscalePolicy
type VipAutoscalePolicy struct {

	// The amount of time, in seconds, when a Vip is withdrawn before a scaling activity starts. Field introduced in 17.2.12, 18.1.2.
	DNSCooldown *int32 `json:"dns_cooldown,omitempty"`

	// The maximum size of the group. Field introduced in 17.2.12, 18.1.2.
	MaxSize *int32 `json:"max_size,omitempty"`

	// The minimum size of the group. Field introduced in 17.2.12, 18.1.2.
	MinSize *int32 `json:"min_size,omitempty"`

	// When set, scaling is suspended. Field introduced in 17.2.12, 18.1.2.
	Suspend *bool `json:"suspend,omitempty"`
}
