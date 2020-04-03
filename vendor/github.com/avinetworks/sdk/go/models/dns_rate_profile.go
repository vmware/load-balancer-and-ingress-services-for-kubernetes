package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DNSRateProfile Dns rate profile
// swagger:model DnsRateProfile
type DNSRateProfile struct {

	// Action to perform upon rate limiting. Field introduced in 18.2.5.
	// Required: true
	Action *DNSRuleRLAction `json:"action"`

	// Maximum number of connections or requests or packets to be rate limited instantaneously. Field introduced in 18.2.5.
	BurstSize *int32 `json:"burst_size,omitempty"`

	// Maximum number of connections or requests or packets per second. Allowed values are 1-4294967295. Special values are 0- 'unlimited'. Field introduced in 18.2.5.
	Count *int32 `json:"count,omitempty"`

	// Enable fine granularity. Field introduced in 18.2.5.
	FineGrain *bool `json:"fine_grain,omitempty"`

	// Time value in seconds to enforce rate count. Allowed values are 1-300. Field introduced in 18.2.5.
	Period *int32 `json:"period,omitempty"`
}
