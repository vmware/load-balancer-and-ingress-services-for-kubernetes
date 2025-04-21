// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ParamInURI param in URI
// swagger:model ParamInURI
type ParamInURI struct {

	// Param name in hitted signature rule match_element. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ParamName *string `json:"param_name,omitempty"`

	// Param value in hitted signature rule match_element. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}
