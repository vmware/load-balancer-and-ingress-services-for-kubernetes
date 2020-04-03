package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RateProfile rate profile
// swagger:model RateProfile
type RateProfile struct {

	// Action to perform upon rate limiting.
	// Required: true
	Action *RateLimiterAction `json:"action"`

	// Maximum number of connections or requests or packets to be let through instantaneously. Allowed values are 10-2500. Special values are 0- 'automatic'.
	BurstSz *int32 `json:"burst_sz,omitempty"`

	// Maximum number of connections or requests or packets. Allowed values are 1-1000000000. Special values are 0- 'unlimited'.
	Count *int32 `json:"count,omitempty"`

	// Explicitly tracks an attacker across rate periods.
	ExplicitTracking *bool `json:"explicit_tracking,omitempty"`

	// Enable fine granularity.
	FineGrain *bool `json:"fine_grain,omitempty"`

	// HTTP cookie name. Field introduced in 17.1.1.
	HTTPCookie *string `json:"http_cookie,omitempty"`

	// HTTP header name. Field introduced in 17.1.1.
	HTTPHeader *string `json:"http_header,omitempty"`

	// Time value in seconds to enforce rate count. Allowed values are 1-300.
	Period *int32 `json:"period,omitempty"`
}
