// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtDatastores nsxt datastores
// swagger:model NsxtDatastores
type NsxtDatastores struct {

	// List of shared datastores. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	DsIds []string `json:"ds_ids,omitempty"`

	// Include or Exclude. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Include *bool `json:"include,omitempty"`
}
