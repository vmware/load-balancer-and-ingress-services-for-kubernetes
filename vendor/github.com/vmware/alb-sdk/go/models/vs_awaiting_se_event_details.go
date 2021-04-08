package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsAwaitingSeEventDetails vs awaiting se event details
// swagger:model VsAwaitingSeEventDetails
type VsAwaitingSeEventDetails struct {

	// Number of awaitingse_timeout.
	// Required: true
	AwaitingseTimeout *int32 `json:"awaitingse_timeout"`

	// ip of VsAwaitingSeEventDetails.
	IP *string `json:"ip,omitempty"`

	// Placeholder for description of property se_assigned of obj type VsAwaitingSeEventDetails field type str  type object
	SeAssigned []*VipSeAssigned `json:"se_assigned,omitempty"`

	// Placeholder for description of property se_requested of obj type VsAwaitingSeEventDetails field type str  type object
	SeRequested *VirtualServiceResource `json:"se_requested,omitempty"`

	// Unique object identifier of vs.
	// Required: true
	VsUUID *string `json:"vs_uuid"`
}
