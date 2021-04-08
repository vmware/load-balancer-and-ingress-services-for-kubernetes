package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeUpgradeMigrateEventDetails se upgrade migrate event details
// swagger:model SeUpgradeMigrateEventDetails
type SeUpgradeMigrateEventDetails struct {

	// Placeholder for description of property migrate_params of obj type SeUpgradeMigrateEventDetails field type str  type object
	MigrateParams *VsMigrateParams `json:"migrate_params,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}
