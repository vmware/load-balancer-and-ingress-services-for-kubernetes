// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PrimaryPool primary pool
// swagger:model PrimaryPool
type PrimaryPool struct {

	// Pool's ID. Field introduced in 20.1.7, 21.1.2, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	PoolUUID *string `json:"pool_uuid"`
}
