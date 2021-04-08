package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CCScaleSetNotifDetails c c scale set notif details
// swagger:model CCScaleSetNotifDetails
type CCScaleSetNotifDetails struct {

	// Cloud id. Field introduced in 18.2.5.
	CcID *string `json:"cc_id,omitempty"`

	// Detailed reason for the scale set notification. Field introduced in 18.2.5.
	Reason *string `json:"reason,omitempty"`

	// Names of scale sets for which polling failed. Field introduced in 18.2.5.
	ScalesetNames []string `json:"scaleset_names,omitempty"`
}
