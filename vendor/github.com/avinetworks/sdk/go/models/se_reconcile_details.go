package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeReconcileDetails se reconcile details
// swagger:model SeReconcileDetails
type SeReconcileDetails struct {

	// Placeholder for description of property new_service_cores of obj type SeReconcileDetails field type str  type number
	NewServiceCores *float64 `json:"new_service_cores,omitempty"`

	// Placeholder for description of property old_service_cores of obj type SeReconcileDetails field type str  type number
	OldServiceCores *float64 `json:"old_service_cores,omitempty"`

	// Unique object identifier of se.
	SeUUID *string `json:"se_uuid,omitempty"`

	// Unique object identifier of tenant.
	TenantUUID *string `json:"tenant_uuid,omitempty"`

	// tier of SeReconcileDetails.
	Tier *string `json:"tier,omitempty"`
}
