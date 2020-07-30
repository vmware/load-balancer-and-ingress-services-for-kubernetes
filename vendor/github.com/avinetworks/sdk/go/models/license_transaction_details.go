package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LicenseTransactionDetails license transaction details
// swagger:model LicenseTransactionDetails
type LicenseTransactionDetails struct {

	// cookie of LicenseTransactionDetails.
	Cookie *string `json:"cookie,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// id of LicenseTransactionDetails.
	ID *string `json:"id,omitempty"`

	// Placeholder for description of property licensed_service_cores of obj type LicenseTransactionDetails field type str  type number
	LicensedServiceCores *float64 `json:"licensed_service_cores,omitempty"`

	// operation of LicenseTransactionDetails.
	Operation *string `json:"operation,omitempty"`

	// Placeholder for description of property overdraft of obj type LicenseTransactionDetails field type str  type boolean
	Overdraft *bool `json:"overdraft,omitempty"`

	// Placeholder for description of property service_cores of obj type LicenseTransactionDetails field type str  type number
	ServiceCores *float64 `json:"service_cores,omitempty"`

	// Unique object identifier of tenant.
	TenantUUID *string `json:"tenant_uuid,omitempty"`

	// tier of LicenseTransactionDetails.
	Tier *string `json:"tier,omitempty"`
}
