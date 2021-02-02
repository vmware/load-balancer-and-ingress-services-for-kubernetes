package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeVsPktBufHighEventDetails se vs pkt buf high event details
// swagger:model SeVsPktBufHighEventDetails
type SeVsPktBufHighEventDetails struct {

	// Current packet buffer usage of the VS.
	CurrentValue *int32 `json:"current_value,omitempty"`

	// Buffer usage threshold value.
	Threshold *int32 `json:"threshold,omitempty"`

	// Virtual Service name. It is a reference to an object of type VirtualService.
	VirtualService *string `json:"virtual_service,omitempty"`
}
