package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeHbRecoveredEventDetails se hb recovered event details
// swagger:model SeHbRecoveredEventDetails
type SeHbRecoveredEventDetails struct {

	// Heartbeat Request/Response received.
	HbType *int32 `json:"hb_type,omitempty"`

	// UUID of the remote SE with which dataplane heartbeat recovered. It is a reference to an object of type ServiceEngine. Field introduced in 20.1.1.
	RemoteSeRef *string `json:"remote_se_ref,omitempty"`

	// UUID of the SE reporting this event. It is a reference to an object of type ServiceEngine. Field introduced in 20.1.1.
	ReportingSeRef *string `json:"reporting_se_ref,omitempty"`

	// UUID of a VS which is placed on reporting-SE and remote-SE. Field introduced in 20.1.1.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
