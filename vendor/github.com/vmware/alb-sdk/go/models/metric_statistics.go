// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricStatistics metric statistics
// swagger:model MetricStatistics
type MetricStatistics struct {

	// value of the last sample.
	LastSample *float64 `json:"last_sample,omitempty"`

	// maximum value in time series requested.
	Max *float64 `json:"max,omitempty"`

	// timestamp of the minimum value.
	MaxTs *string `json:"max_ts,omitempty"`

	// arithmetic mean.
	Mean *float64 `json:"mean,omitempty"`

	// minimum value in time series requested.
	Min *float64 `json:"min,omitempty"`

	// timestamp of the minimum value.
	MinTs *string `json:"min_ts,omitempty"`

	// Number of actual data samples. It excludes fake data.
	NumSamples *int32 `json:"num_samples,omitempty"`

	// summation of all values.
	Sum *float64 `json:"sum,omitempty"`

	// slope of the data points across the time series requested. trend = (last_value - avg)/avg.
	Trend *float64 `json:"trend,omitempty"`
}
