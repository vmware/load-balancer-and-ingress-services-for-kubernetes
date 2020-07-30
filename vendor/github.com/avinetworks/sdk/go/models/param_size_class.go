package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ParamSizeClass param size class
// swagger:model ParamSizeClass
type ParamSizeClass struct {

	// Number of hits.
	Hits *int64 `json:"hits,omitempty"`

	//  Enum options - EMPTY, SMALL, MEDIUM, LARGE, UNLIMITED.
	Len *string `json:"len,omitempty"`
}
