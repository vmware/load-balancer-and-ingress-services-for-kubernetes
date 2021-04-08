package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsScaleinParams vs scalein params
// swagger:model VsScaleinParams
type VsScaleinParams struct {

	// Placeholder for description of property admin_down of obj type VsScaleinParams field type str  type boolean
	AdminDown *bool `json:"admin_down,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	FromSeRef *string `json:"from_se_ref,omitempty"`

	// Placeholder for description of property scalein_primary of obj type VsScaleinParams field type str  type boolean
	ScaleinPrimary *bool `json:"scalein_primary,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	//  Field introduced in 17.1.1.
	// Required: true
	VipID *string `json:"vip_id"`
}
