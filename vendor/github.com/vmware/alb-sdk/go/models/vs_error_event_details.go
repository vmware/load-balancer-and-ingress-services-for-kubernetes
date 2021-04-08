package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsErrorEventDetails vs error event details
// swagger:model VsErrorEventDetails
type VsErrorEventDetails struct {

	// error_message of VsErrorEventDetails.
	ErrorMessage *string `json:"error_message,omitempty"`

	// ip of VsErrorEventDetails.
	IP *string `json:"ip,omitempty"`

	// ip6 of VsErrorEventDetails.
	Ip6 *string `json:"ip6,omitempty"`

	// Number of rpc_status.
	RPCStatus *int64 `json:"rpc_status,omitempty"`

	// Placeholder for description of property se_assigned of obj type VsErrorEventDetails field type str  type object
	SeAssigned []*VipSeAssigned `json:"se_assigned,omitempty"`

	// Placeholder for description of property se_requested of obj type VsErrorEventDetails field type str  type object
	SeRequested *VirtualServiceResource `json:"se_requested,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}
