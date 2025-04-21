// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ParamTypeClass param type class
// swagger:model ParamTypeClass
type ParamTypeClass struct {

	//  Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hits uint64 `json:"hits,omitempty"`

	//  Enum options - PARAM_FLAG, PARAM_DIGITS, PARAM_HEXDIGITS, PARAM_WORD, PARAM_SAFE_TEXT, PARAM_SAFE_TEXT_MULTILINE, PARAM_TEXT, PARAM_TEXT_MULTILINE, PARAM_ALL. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`
}
