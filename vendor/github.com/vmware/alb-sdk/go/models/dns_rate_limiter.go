package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRateLimiter Dns rate limiter
// swagger:model DnsRateLimiter
type DNSRateLimiter struct {

	// Action to perform upon rate limiting. Field introduced in 20.1.1.
	// Required: true
	Action *DNSRuleRLAction `json:"action"`

	// Rate limiting object. Field introduced in 20.1.1.
	// Required: true
	RateLimiterObject *RateLimiter `json:"rate_limiter_object"`
}
