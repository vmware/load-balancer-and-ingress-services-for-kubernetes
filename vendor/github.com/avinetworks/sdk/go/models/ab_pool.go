package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AbPool ab pool
// swagger:model AbPool
type AbPool struct {

	// Pool configured as B pool for A/B testing. It is a reference to an object of type Pool.
	// Required: true
	PoolRef *string `json:"pool_ref"`

	// Ratio of traffic diverted to the B pool, for A/B testing. Allowed values are 0-100.
	Ratio *int32 `json:"ratio,omitempty"`
}
