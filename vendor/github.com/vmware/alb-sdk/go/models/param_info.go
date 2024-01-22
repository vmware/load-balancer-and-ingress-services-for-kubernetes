// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ParamInfo param info
// swagger:model ParamInfo
type ParamInfo struct {

	// Number of hits for a param. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ParamHits uint64 `json:"param_hits,omitempty"`

	// Param name. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ParamKey *string `json:"param_key,omitempty"`

	// Various param size and its respective hit count. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ParamSizeClasses []*ParamSizeClass `json:"param_size_classes,omitempty"`

	// Various param type and its respective hit count. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ParamTypeClasses []*ParamTypeClass `json:"param_type_classes,omitempty"`
}
