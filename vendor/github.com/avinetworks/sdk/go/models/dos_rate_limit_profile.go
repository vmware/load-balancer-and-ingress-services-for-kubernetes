package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DosRateLimitProfile dos rate limit profile
// swagger:model DosRateLimitProfile
type DosRateLimitProfile struct {

	// Profile for DoS attack detection.
	DosProfile *DosThresholdProfile `json:"dos_profile,omitempty"`

	// Profile for Connections/Requests rate limiting.
	RlProfile *RateLimiterProfile `json:"rl_profile,omitempty"`
}
