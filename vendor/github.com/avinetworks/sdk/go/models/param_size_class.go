package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ParamSizeClass param size class
// swagger:model ParamSizeClass
type ParamSizeClass struct {

	//  Field introduced in 20.1.1.
	Hits *int64 `json:"hits,omitempty"`

	//  Enum options - EMPTY, SMALL, MEDIUM, LARGE, UNLIMITED. Field introduced in 20.1.1.
	Len *string `json:"len,omitempty"`
}
