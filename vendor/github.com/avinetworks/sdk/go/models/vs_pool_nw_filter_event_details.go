package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsPoolNwFilterEventDetails vs pool nw filter event details
// swagger:model VsPoolNwFilterEventDetails
type VsPoolNwFilterEventDetails struct {

	// filter of VsPoolNwFilterEventDetails.
	// Required: true
	Filter *string `json:"filter"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// network of VsPoolNwFilterEventDetails.
	// Required: true
	Network *string `json:"network"`
}
