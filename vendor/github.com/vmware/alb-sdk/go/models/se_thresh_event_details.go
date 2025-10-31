// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeThreshEventDetails se thresh event details
// swagger:model SeThreshEventDetails
type SeThreshEventDetails struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	CurrValue *uint64 `json:"curr_value"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Thresh *uint64 `json:"thresh"`
}
