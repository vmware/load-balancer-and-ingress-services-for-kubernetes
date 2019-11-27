package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeHBEventDetails se h b event details
// swagger:model SeHBEventDetails
type SeHBEventDetails struct {

	// HB Request/Response not received.
	HbType *int32 `json:"hb_type,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef1 *string `json:"se_ref1,omitempty"`

	// UUID of a SE in the SE-Group which failed to respond. It is a reference to an object of type ServiceEngine.
	SeRef2 *string `json:"se_ref2,omitempty"`
}
