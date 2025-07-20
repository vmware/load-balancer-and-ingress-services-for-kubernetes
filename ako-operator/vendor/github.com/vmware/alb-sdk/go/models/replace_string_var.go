// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ReplaceStringVar replace *string var
// swagger:model ReplaceStringVar
type ReplaceStringVar struct {

	// Type of replacement *string - can be a variable exposed from datascript, value of an HTTP variable, a custom user-input literal string, or a *string with all three combined. Enum options - DATASCRIPT_VAR, AVI_VAR, LITERAL_STRING, COMBINATION_STRING. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`

	// Value of the replacement *string - name of variable exposed from datascript, name of the HTTP header, a custom user-input literal string, or a *string with all three combined. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Val *string `json:"val,omitempty"`
}
