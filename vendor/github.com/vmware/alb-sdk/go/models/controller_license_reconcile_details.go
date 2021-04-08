package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerLicenseReconcileDetails controller license reconcile details
// swagger:model ControllerLicenseReconcileDetails
type ControllerLicenseReconcileDetails struct {

	// Placeholder for description of property new_available_service_cores of obj type ControllerLicenseReconcileDetails field type str  type number
	NewAvailableServiceCores *float64 `json:"new_available_service_cores,omitempty"`

	// Placeholder for description of property new_consumed_service_cores of obj type ControllerLicenseReconcileDetails field type str  type number
	NewConsumedServiceCores *float64 `json:"new_consumed_service_cores,omitempty"`

	// Placeholder for description of property new_escrow_service_cores of obj type ControllerLicenseReconcileDetails field type str  type number
	NewEscrowServiceCores *float64 `json:"new_escrow_service_cores,omitempty"`

	// Placeholder for description of property new_remaining_service_cores of obj type ControllerLicenseReconcileDetails field type str  type number
	NewRemainingServiceCores *float64 `json:"new_remaining_service_cores,omitempty"`

	// Placeholder for description of property old_available_service_cores of obj type ControllerLicenseReconcileDetails field type str  type number
	OldAvailableServiceCores *float64 `json:"old_available_service_cores,omitempty"`

	// Placeholder for description of property old_consumed_service_cores of obj type ControllerLicenseReconcileDetails field type str  type number
	OldConsumedServiceCores *float64 `json:"old_consumed_service_cores,omitempty"`

	// Placeholder for description of property old_escrow_service_cores of obj type ControllerLicenseReconcileDetails field type str  type number
	OldEscrowServiceCores *float64 `json:"old_escrow_service_cores,omitempty"`

	// Placeholder for description of property old_remaining_service_cores of obj type ControllerLicenseReconcileDetails field type str  type number
	OldRemainingServiceCores *float64 `json:"old_remaining_service_cores,omitempty"`

	// Unique object identifier of tenant.
	TenantUUID *string `json:"tenant_uuid,omitempty"`

	// tier of ControllerLicenseReconcileDetails.
	Tier *string `json:"tier,omitempty"`
}
