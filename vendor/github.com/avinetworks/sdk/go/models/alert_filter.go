package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertFilter alert filter
// swagger:model AlertFilter
type AlertFilter struct {

	// filter_action of AlertFilter.
	FilterAction *string `json:"filter_action,omitempty"`

	// filter_string of AlertFilter.
	// Required: true
	FilterString *string `json:"filter_string"`
}
