// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Selector selector
// swagger:model Selector
type Selector struct {

	// Labels as key value pairs to select on. Field introduced in 20.1.3. Minimum of 1 items required.
	Labels []*KeyValueTuple `json:"labels,omitempty"`

	// Selector type. Enum options - SELECTOR_IPAM. Field introduced in 20.1.3.
	// Required: true
	Type *string `json:"type"`
}
