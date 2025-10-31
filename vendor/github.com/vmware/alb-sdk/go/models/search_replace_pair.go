// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SearchReplacePair search replace pair
// swagger:model SearchReplacePair
type SearchReplacePair struct {

	// String to replace the searched value. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ReplacementString *ReplaceStringVar `json:"replacement_string,omitempty"`

	// String to search for in the body. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	SearchString *SearchStringVar `json:"search_string"`
}
