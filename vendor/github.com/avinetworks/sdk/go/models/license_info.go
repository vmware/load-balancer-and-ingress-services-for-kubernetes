package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LicenseInfo license info
// swagger:model LicenseInfo
type LicenseInfo struct {

	// Last updated time. Field introduced in 20.1.1.
	// Required: true
	LastUpdated *int64 `json:"last_updated"`

	// Quantity of service cores. Field introduced in 20.1.1.
	// Required: true
	ServiceCores *float64 `json:"service_cores"`

	// Specifies the license tier. Field introduced in 20.1.1.
	TenantUUID *string `json:"tenant_uuid,omitempty"`

	// Specifies the license tier. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC. Field introduced in 20.1.1.
	// Required: true
	Tier *string `json:"tier"`

	// Identifier(license_id, se_uuid, cookie). Field introduced in 20.1.1.
	UUID *string `json:"uuid,omitempty"`
}
