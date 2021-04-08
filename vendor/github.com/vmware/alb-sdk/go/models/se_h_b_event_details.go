package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeHBEventDetails se h b event details
// swagger:model SeHBEventDetails
type SeHBEventDetails struct {

	// HB Request/Response not received.
	HbType *int32 `json:"hb_type,omitempty"`

	// UUID of the SE with which Heartbeat failed. It is a reference to an object of type ServiceEngine. Field introduced in 20.1.1.
	RemoteSeRef *string `json:"remote_se_ref,omitempty"`

	// UUID of the SE reporting this event. It is a reference to an object of type ServiceEngine. Field introduced in 20.1.1.
	ReportingSeRef *string `json:"reporting_se_ref,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Field deprecated in 20.1.1.
	SeRef1 *string `json:"se_ref1,omitempty"`

	// UUID of a SE in the SE-Group which failed to respond. It is a reference to an object of type ServiceEngine. Field deprecated in 20.1.1.
	SeRef2 *string `json:"se_ref2,omitempty"`

	// UUID of the virtual service which is placed on reporting-SE and remote-SE. Field introduced in 20.1.1.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
