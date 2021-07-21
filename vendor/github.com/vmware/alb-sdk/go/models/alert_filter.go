// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertFilter alert filter
// swagger:model AlertFilter
type AlertFilter struct {

	// filter_action of AlertFilter.
	FilterAction *string `json:"filter_action,omitempty"`

	// filter_string of AlertFilter.
	// Required: true
	FilterString *string `json:"filter_string"`
}
