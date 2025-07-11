// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// InventoryMetrics inventory metrics
// swagger:model InventoryMetrics
type InventoryMetrics struct {

	// Metric data. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Data *InventoryMetricsData `json:"data,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Headers *InventoryMetricsHeaders `json:"headers,omitempty"`
}
