// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDbDiskEventDetails metrics db disk event details
// swagger:model MetricsDbDiskEventDetails
type MetricsDbDiskEventDetails struct {

	// List of dropped metrics tables. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsDeletedTables []string `json:"metrics_deleted_tables,omitempty"`

	// Total size of freed metrics tables. In KBs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MetricsFreeSz *float64 `json:"metrics_free_sz"`

	// Disk quota allocated for Metrics DB. In GBs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MetricsQuota *float64 `json:"metrics_quota"`
}
