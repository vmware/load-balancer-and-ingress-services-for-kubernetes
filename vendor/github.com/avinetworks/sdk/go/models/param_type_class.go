package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ParamTypeClass param type class
// swagger:model ParamTypeClass
type ParamTypeClass struct {

	//  Field introduced in 20.1.1.
	Hits *int64 `json:"hits,omitempty"`

	//  Enum options - PARAM_FLAG, PARAM_DIGITS, PARAM_HEXDIGITS, PARAM_WORD, PARAM_SAFE_TEXT, PARAM_SAFE_TEXT_MULTILINE, PARAM_TEXT, PARAM_TEXT_MULTILINE, PARAM_ALL. Field introduced in 20.1.1.
	Type *string `json:"type,omitempty"`
}
