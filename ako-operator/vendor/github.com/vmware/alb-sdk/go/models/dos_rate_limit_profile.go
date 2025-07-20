// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DosRateLimitProfile dos rate limit profile
// swagger:model DosRateLimitProfile
type DosRateLimitProfile struct {

	// Profile for DoS attack detection. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DosProfile *DosThresholdProfile `json:"dos_profile,omitempty"`

	// Profile for Connections/Requests rate limiting. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RlProfile *RateLimiterProfile `json:"rl_profile,omitempty"`
}
