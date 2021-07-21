// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeThreshEventDetails se thresh event details
// swagger:model SeThreshEventDetails
type SeThreshEventDetails struct {

	// Number of curr_value.
	// Required: true
	CurrValue *int64 `json:"curr_value"`

	// Number of thresh.
	// Required: true
	Thresh *int64 `json:"thresh"`
}
