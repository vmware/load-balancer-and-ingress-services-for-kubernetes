package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServerScaleOutParams server scale out params
// swagger:model ServerScaleOutParams
type ServerScaleOutParams struct {

	// Reason for the manual scaleout.
	Reason *string `json:"reason,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
