package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RateLimiter rate limiter
// swagger:model RateLimiter
type RateLimiter struct {

	// Maximum number of connections, requests or packets to be let through instantaneously.  If this is less than count, it will have no effect. Allowed values are 0-1000000000. Field introduced in 18.2.9.
	BurstSz *int32 `json:"burst_sz,omitempty"`

	// Maximum number of connections, requests or packets permitted each period. Allowed values are 1-1000000000. Field introduced in 18.2.9.
	// Required: true
	Count *int32 `json:"count"`

	// Identifier for Rate Limit. Constructed according to context. Field introduced in 18.2.9.
	Name *string `json:"name,omitempty"`

	// Time value in seconds to enforce rate count. Allowed values are 1-1000000000. Field introduced in 18.2.9. Unit is SEC.
	// Required: true
	Period *int32 `json:"period"`
}
