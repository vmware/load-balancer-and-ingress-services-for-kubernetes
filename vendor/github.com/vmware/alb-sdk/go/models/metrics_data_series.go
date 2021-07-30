// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDataSeries metrics data series
// swagger:model MetricsDataSeries
type MetricsDataSeries struct {

	// Placeholder for description of property data of obj type MetricsDataSeries field type str  type object
	Data []*MetricsData `json:"data,omitempty"`

	// Placeholder for description of property header of obj type MetricsDataSeries field type str  type object
	// Required: true
	Header *MetricsDataHeader `json:"header"`
}
