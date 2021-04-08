package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SwitchoverEventDetails switchover event details
// swagger:model SwitchoverEventDetails
type SwitchoverEventDetails struct {

	// from_se_name of SwitchoverEventDetails.
	FromSeName *string `json:"from_se_name,omitempty"`

	// ip of SwitchoverEventDetails.
	IP *string `json:"ip,omitempty"`

	// ip6 of SwitchoverEventDetails.
	Ip6 *string `json:"ip6,omitempty"`

	// to_se_name of SwitchoverEventDetails.
	ToSeName *string `json:"to_se_name,omitempty"`

	// vs_name of SwitchoverEventDetails.
	VsName *string `json:"vs_name,omitempty"`

	// Unique object identifier of vs.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
