// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// InventoryConfig inventory config
// swagger:model InventoryConfig
type InventoryConfig struct {

	// Allow inventory stats to be regularly sent to Pulse Cloud Services. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enable *bool `json:"enable,omitempty"`
}
