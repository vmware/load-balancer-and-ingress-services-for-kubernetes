// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LicenseTierUsage license tier usage
// swagger:model LicenseTierUsage
type LicenseTierUsage struct {

	// Specifies the license tier. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS. Field introduced in 20.1.1.
	Tier *string `json:"tier,omitempty"`

	// Usage stats of license tier. Field introduced in 20.1.1.
	Usage *LicenseUsage `json:"usage,omitempty"`
}
