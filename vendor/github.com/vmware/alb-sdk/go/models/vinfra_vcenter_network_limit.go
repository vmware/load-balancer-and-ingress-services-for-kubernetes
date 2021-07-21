// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VinfraVcenterNetworkLimit vinfra vcenter network limit
// swagger:model VinfraVcenterNetworkLimit
type VinfraVcenterNetworkLimit struct {

	// additional_reason of VinfraVcenterNetworkLimit.
	// Required: true
	AdditionalReason *string `json:"additional_reason"`

	// Number of current.
	// Required: true
	Current *int64 `json:"current"`

	// Number of limit.
	// Required: true
	Limit *int64 `json:"limit"`
}
