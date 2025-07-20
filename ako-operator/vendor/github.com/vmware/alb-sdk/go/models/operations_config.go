// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OperationsConfig operations config
// swagger:model OperationsConfig
type OperationsConfig struct {

	// Inventory op config. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	InventoryConfig *InventoryConfig `json:"inventory_config,omitempty"`
}
