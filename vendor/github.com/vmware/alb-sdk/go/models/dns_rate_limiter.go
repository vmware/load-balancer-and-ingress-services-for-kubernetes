// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRateLimiter Dns rate limiter
// swagger:model DnsRateLimiter
type DNSRateLimiter struct {

	// Action to perform upon rate limiting. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Action *DNSRuleRLAction `json:"action"`

	// Rate limiting object. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	RateLimiterObject *RateLimiter `json:"rate_limiter_object"`
}
