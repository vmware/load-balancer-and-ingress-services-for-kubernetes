package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LicenseTierUsage license tier usage
// swagger:model LicenseTierUsage
type LicenseTierUsage struct {

	// Specifies the license tier. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS. Field introduced in 20.1.1.
	Tier *string `json:"tier,omitempty"`

	// Usage stats of license tier. Field introduced in 20.1.1.
	Usage *LicenseUsage `json:"usage,omitempty"`
}
