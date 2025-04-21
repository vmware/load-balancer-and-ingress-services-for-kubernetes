// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// InventoryMetricStatistics inventory metric statistics
// swagger:model InventoryMetricStatistics
type InventoryMetricStatistics struct {

	// Maximum value in time series requested. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Max *float64 `json:"max,omitempty"`

	// Arithmetic mean. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Mean *float64 `json:"mean,omitempty"`

	// Minimum value in time series requested. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Min *float64 `json:"min,omitempty"`

	// Number of actual data samples. It excludes fake data. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSamples uint32 `json:"num_samples,omitempty"`
}
