package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HttpsecurityActionRateProfile httpsecurity action rate profile
// swagger:model HTTPSecurityActionRateProfile
type HttpsecurityActionRateProfile struct {

	// The action to take when the rate limit has been reached. Field introduced in 18.2.9.
	// Required: true
	Action *RateLimiterAction `json:"action"`

	// Rate limiting should be done on a per client ip basis. Field introduced in 18.2.9.
	PerClientIP *bool `json:"per_client_ip,omitempty"`

	// Rate limiting should be done on a per request uri path basis. Field introduced in 18.2.9.
	PerURIPath *bool `json:"per_uri_path,omitempty"`

	// The rate limiter used when this action is triggered. Field introduced in 18.2.9.
	// Required: true
	RateLimiter *RateLimiter `json:"rate_limiter"`
}
