// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ParamSizeClass param size class
// swagger:model ParamSizeClass
type ParamSizeClass struct {

	//  Field introduced in 20.1.1.
	Hits *int64 `json:"hits,omitempty"`

	//  Enum options - EMPTY, SMALL, MEDIUM, LARGE, UNLIMITED. Field introduced in 20.1.1.
	Len *string `json:"len,omitempty"`
}
