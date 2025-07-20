// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SearchStringVar search *string var
// swagger:model SearchStringVar
type SearchStringVar struct {

	// Type of search *string - can be a variable exposed from datascript, value of an HTTP variable, a custom user-input literal string, or a regular expression. Enum options - SEARCH_DATASCRIPT_VAR, SEARCH_AVI_VAR, SEARCH_LITERAL_STRING, SEARCH_REGEX. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`

	// Value of search *string - can be a variable exposed from datascript, value of an HTTP variable, a custom user-input literal string, or a regular expression. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Val *string `json:"val"`
}
