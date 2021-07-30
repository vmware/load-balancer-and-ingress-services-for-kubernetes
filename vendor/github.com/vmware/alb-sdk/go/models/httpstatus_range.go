// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpstatusRange httpstatus range
// swagger:model HTTPStatusRange
type HttpstatusRange struct {

	// Starting HTTP response status code.
	// Required: true
	Begin *int32 `json:"begin"`

	// Ending HTTP response status code.
	// Required: true
	End *int32 `json:"end"`
}
