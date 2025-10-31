// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolGroupRefs pool group refs
// swagger:model PoolGroupRefs
type PoolGroupRefs struct {

	// UUID of the pool group. It is a reference to an object of type PoolGroup. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ref *string `json:"ref,omitempty"`
}
