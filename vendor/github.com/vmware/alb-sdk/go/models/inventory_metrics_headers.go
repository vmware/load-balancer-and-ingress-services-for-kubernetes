// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// InventoryMetricsHeaders inventory metrics headers
// swagger:model InventoryMetricsHeaders
type InventoryMetricsHeaders struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Statistics *InventoryMetricStatistics `json:"statistics,omitempty"`
}
