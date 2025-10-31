// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// InventoryConfiguration inventory configuration
// swagger:model InventoryConfiguration
type InventoryConfiguration struct {

	// Names, IP's of VS, Pool(PoolGroup) servers would be searchable on Cloud console. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	EnableSearchInfo *bool `json:"enable_search_info,omitempty"`
}
