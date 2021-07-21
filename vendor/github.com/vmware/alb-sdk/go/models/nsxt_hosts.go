// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtHosts nsxt hosts
// swagger:model NsxtHosts
type NsxtHosts struct {

	// List of transport nodes. Field introduced in 20.1.1.
	HostIds []string `json:"host_ids,omitempty"`

	// Include or Exclude. Field introduced in 20.1.1.
	Include *bool `json:"include,omitempty"`
}
