// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtDatastores nsxt datastores
// swagger:model NsxtDatastores
type NsxtDatastores struct {

	// List of shared datastores. Field introduced in 20.1.2. Allowed in Basic edition, Enterprise edition.
	DsIds []string `json:"ds_ids,omitempty"`

	// Include or Exclude. Field introduced in 20.1.2. Allowed in Basic edition, Enterprise edition.
	Include *bool `json:"include,omitempty"`
}
