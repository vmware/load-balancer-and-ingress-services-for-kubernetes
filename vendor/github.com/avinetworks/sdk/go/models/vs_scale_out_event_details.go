package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsScaleOutEventDetails vs scale out event details
// swagger:model VsScaleOutEventDetails
type VsScaleOutEventDetails struct {

	// error_message of VsScaleOutEventDetails.
	ErrorMessage *string `json:"error_message,omitempty"`

	// ip of VsScaleOutEventDetails.
	IP *string `json:"ip,omitempty"`

	// ip6 of VsScaleOutEventDetails.
	Ip6 *string `json:"ip6,omitempty"`

	// Number of rpc_status.
	RPCStatus *int64 `json:"rpc_status,omitempty"`

	// Placeholder for description of property scale_status of obj type VsScaleOutEventDetails field type str  type object
	ScaleStatus *ScaleStatus `json:"scale_status,omitempty"`

	// Placeholder for description of property se_assigned of obj type VsScaleOutEventDetails field type str  type object
	SeAssigned []*VipSeAssigned `json:"se_assigned,omitempty"`

	// Placeholder for description of property se_requested of obj type VsScaleOutEventDetails field type str  type object
	SeRequested *VirtualServiceResource `json:"se_requested,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}
