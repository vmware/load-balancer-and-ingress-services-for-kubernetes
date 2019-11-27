package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DisableSeMigrateEventDetails disable se migrate event details
// swagger:model DisableSeMigrateEventDetails
type DisableSeMigrateEventDetails struct {

	// Placeholder for description of property migrate_params of obj type DisableSeMigrateEventDetails field type str  type object
	MigrateParams *VsMigrateParams `json:"migrate_params,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}
