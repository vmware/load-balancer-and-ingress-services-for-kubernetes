package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsFsmEventDetails vs fsm event details
// swagger:model VsFsmEventDetails
type VsFsmEventDetails struct {

	// vip_id of VsFsmEventDetails.
	VipID *string `json:"vip_id,omitempty"`

	// Placeholder for description of property vs_rt of obj type VsFsmEventDetails field type str  type object
	VsRt *VirtualServiceRuntime `json:"vs_rt,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}
