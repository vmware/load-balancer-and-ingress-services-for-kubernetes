// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// EventCache event cache
// swagger:model EventCache
type EventCache struct {

	// Placeholder for description of property dns_state of obj type EventCache field type str  type boolean
	DNSState *bool `json:"dns_state,omitempty"`

	// Cache the exception strings in the system.
	Exceptions []string `json:"exceptions,omitempty"`
}
