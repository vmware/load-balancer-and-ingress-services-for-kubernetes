// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LogManagerDebugFilter log manager debug filter
// swagger:model LogManagerDebugFilter
type LogManagerDebugFilter struct {

	// UUID of the entity. It is a reference to an object of type Virtualservice. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EntityRef *string `json:"entity_ref,omitempty"`
}
