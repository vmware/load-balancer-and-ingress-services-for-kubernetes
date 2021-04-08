package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsInitialPlacementEventDetails vs initial placement event details
// swagger:model VsInitialPlacementEventDetails
type VsInitialPlacementEventDetails struct {

	// error_message of VsInitialPlacementEventDetails.
	ErrorMessage *string `json:"error_message,omitempty"`

	// ip of VsInitialPlacementEventDetails.
	IP *string `json:"ip,omitempty"`

	// Number of rpc_status.
	RPCStatus *int64 `json:"rpc_status,omitempty"`

	// Placeholder for description of property se_assigned of obj type VsInitialPlacementEventDetails field type str  type object
	SeAssigned []*VipSeAssigned `json:"se_assigned,omitempty"`

	// Placeholder for description of property se_requested of obj type VsInitialPlacementEventDetails field type str  type object
	SeRequested *VirtualServiceResource `json:"se_requested,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}
