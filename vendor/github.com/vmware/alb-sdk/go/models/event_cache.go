// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// EventCache event cache
// swagger:model EventCache
type EventCache struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSState *bool `json:"dns_state,omitempty"`

	// Cache the exception strings in the system. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Exceptions []string `json:"exceptions,omitempty"`
}
