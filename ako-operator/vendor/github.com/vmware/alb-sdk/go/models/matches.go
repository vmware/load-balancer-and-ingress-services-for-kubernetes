// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Matches matches
// swagger:model Matches
type Matches struct {

	// Matches in signature rule. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchElement *string `json:"match_element,omitempty"`

	// Match value in signature rule. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchValue *string `json:"match_value,omitempty"`
}
