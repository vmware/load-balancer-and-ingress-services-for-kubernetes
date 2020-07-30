package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ParamInfo param info
// swagger:model ParamInfo
type ParamInfo struct {

	// Number of hits for a param. Field introduced in 20.1.1.
	ParamHits *int64 `json:"param_hits,omitempty"`

	// Param name. Field introduced in 20.1.1.
	ParamKey *string `json:"param_key,omitempty"`

	// Various param size and its respective hit count. Field introduced in 20.1.1.
	ParamSizeClasses []*ParamSizeClass `json:"param_size_classes,omitempty"`

	// Various param type and its respective hit count. Field introduced in 20.1.1.
	ParamTypeClasses []*ParamTypeClass `json:"param_type_classes,omitempty"`
}
