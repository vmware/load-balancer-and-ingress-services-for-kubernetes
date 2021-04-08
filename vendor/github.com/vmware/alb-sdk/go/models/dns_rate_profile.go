package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRateProfile Dns rate profile
// swagger:model DnsRateProfile
type DNSRateProfile struct {

	// Action to perform upon rate limiting. Field deprecated in 20.1.1. Field introduced in 18.2.5.
	// Required: true
	Action *DNSRuleRLAction `json:"action"`

	// Maximum number of connections or requests or packets to be rate limited instantaneously. Field deprecated in 20.1.1. Field introduced in 18.2.5.
	BurstSize *int32 `json:"burst_size,omitempty"`

	// Maximum number of connections or requests or packets per second. It is deprecated because of adoption of new shared rate limiter protobuf. Allowed values are 1-4294967295. Special values are 0- 'unlimited'. Field deprecated in 20.1.1. Field introduced in 18.2.5.
	Count *int32 `json:"count,omitempty"`

	// Enable fine granularity. Field deprecated in 20.1.1. Field introduced in 18.2.5.
	FineGrain *bool `json:"fine_grain,omitempty"`

	// Time value in seconds to enforce rate count. Allowed values are 1-300. Field deprecated in 20.1.1. Field introduced in 18.2.5. Unit is SEC.
	Period *int32 `json:"period,omitempty"`
}
