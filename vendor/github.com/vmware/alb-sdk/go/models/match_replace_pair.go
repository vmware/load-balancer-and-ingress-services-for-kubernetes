// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MatchReplacePair match replace pair
// swagger:model MatchReplacePair
type MatchReplacePair struct {

	// String to be matched.
	// Required: true
	MatchString *string `json:"match_string"`

	// Replacement string.
	ReplacementString *ReplaceStringVar `json:"replacement_string,omitempty"`
}
