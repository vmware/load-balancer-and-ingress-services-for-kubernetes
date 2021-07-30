// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LicenseUsage license usage
// swagger:model LicenseUsage
type LicenseUsage struct {

	// Total license cores available for consumption. Field introduced in 20.1.1.
	Available *float64 `json:"available,omitempty"`

	// Total license cores consumed. Field introduced in 20.1.1.
	Consumed *float64 `json:"consumed,omitempty"`

	// Total license cores reserved or escrowed. Field introduced in 20.1.1.
	Escrow *float64 `json:"escrow,omitempty"`

	// Total license cores remaining for consumption. Field introduced in 20.1.1.
	Remaining *float64 `json:"remaining,omitempty"`
}
