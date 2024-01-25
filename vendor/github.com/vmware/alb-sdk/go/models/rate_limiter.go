// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RateLimiter rate limiter
// swagger:model RateLimiter
type RateLimiter struct {

	// Maximum number of connections, requests or packets to be let through instantaneously.  If this is less than count, it will have no effect. Allowed values are 0-1000000000. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Basic edition(Allowed values- 0), Essentials, Enterprise with Cloud Services edition.
	BurstSz uint32 `json:"burst_sz,omitempty"`

	// Maximum number of connections, requests or packets permitted each period. Allowed values are 1-1000000000. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Count *uint32 `json:"count"`

	// Identifier for Rate Limit. Constructed according to context. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Time value in seconds to enforce rate count. Allowed values are 1-1000000000. Field introduced in 18.2.9. Unit is SEC. Allowed in Enterprise edition with any value, Basic edition(Allowed values- 1), Essentials, Enterprise with Cloud Services edition.
	// Required: true
	Period *uint32 `json:"period"`
}
