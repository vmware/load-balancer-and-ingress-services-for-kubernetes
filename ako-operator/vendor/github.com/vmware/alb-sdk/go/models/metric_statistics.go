// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricStatistics metric statistics
// swagger:model MetricStatistics
type MetricStatistics struct {

	// value of the last sample. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LastSample *float64 `json:"last_sample,omitempty"`

	// maximum value in time series requested. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Max *float64 `json:"max,omitempty"`

	// timestamp of the minimum value. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxTs *string `json:"max_ts,omitempty"`

	// arithmetic mean. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Mean *float64 `json:"mean,omitempty"`

	// minimum value in time series requested. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Min *float64 `json:"min,omitempty"`

	// timestamp of the minimum value. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinTs *string `json:"min_ts,omitempty"`

	// Number of actual data samples. It excludes fake data. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumSamples *uint32 `json:"num_samples,omitempty"`

	// summation of all values. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Sum *float64 `json:"sum,omitempty"`

	// slope of the data points across the time series requested. trend = (last_value - avg)/avg. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Trend *float64 `json:"trend,omitempty"`
}
