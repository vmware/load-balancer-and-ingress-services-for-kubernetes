// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDimensionData metrics dimension data
// swagger:model MetricsDimensionData
type MetricsDimensionData struct {

	// Dimension Type. Enum options - METRICS_DIMENSION_METRIC_TIMESTAMP, METRICS_DIMENSION_COUNTRY, METRICS_DIMENSION_OS, METRICS_DIMENSION_URL, METRICS_DIMENSION_DEVTYPE, METRICS_DIMENSION_LANG, METRICS_DIMENSION_BROWSER, METRICS_DIMENSION_IPGROUP, METRICS_DIMENSION_ATTACK, METRICS_DIMENSION_ASN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Dimension *string `json:"dimension"`

	// Dimension ID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	DimensionID *string `json:"dimension_id"`
}
