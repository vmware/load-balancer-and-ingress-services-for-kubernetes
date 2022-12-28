// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpsecurityActionRateProfile httpsecurity action rate profile
// swagger:model HTTPSecurityActionRateProfile
type HttpsecurityActionRateProfile struct {

	// The action to take when the rate limit has been reached. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Action *RateLimiterAction `json:"action"`

	// Rate limiting should be done on a per client ip basis. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PerClientIP *bool `json:"per_client_ip,omitempty"`

	// Rate limiting should be done on a per request uri path basis. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PerURIPath *bool `json:"per_uri_path,omitempty"`

	// The rate limiter used when this action is triggered. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RateLimiter *RateLimiter `json:"rate_limiter"`
}
