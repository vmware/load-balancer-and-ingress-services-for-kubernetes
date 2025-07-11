// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtHosts nsxt hosts
// swagger:model NsxtHosts
type NsxtHosts struct {

	// List of transport nodes. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostIds []string `json:"host_ids,omitempty"`

	// Include or Exclude. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Include *bool `json:"include,omitempty"`
}
