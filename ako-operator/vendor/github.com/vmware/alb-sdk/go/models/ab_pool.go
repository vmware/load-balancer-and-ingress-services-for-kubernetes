// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AbPool ab pool
// swagger:model AbPool
type AbPool struct {

	// Pool configured as B pool for A/B testing. It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	PoolRef *string `json:"pool_ref"`

	// Ratio of traffic diverted to the B pool, for A/B testing. Allowed values are 0-100. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ratio *uint32 `json:"ratio,omitempty"`
}
