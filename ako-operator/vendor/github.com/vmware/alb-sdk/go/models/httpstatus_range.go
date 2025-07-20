// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpstatusRange httpstatus range
// swagger:model HTTPStatusRange
type HttpstatusRange struct {

	// Starting HTTP response status code. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Begin *int32 `json:"begin"`

	// Ending HTTP response status code. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	End *int32 `json:"end"`
}
