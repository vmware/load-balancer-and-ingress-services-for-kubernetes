// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDbDiskEventDetails metrics db disk event details
// swagger:model MetricsDbDiskEventDetails
type MetricsDbDiskEventDetails struct {

	// metrics_deleted_tables of MetricsDbDiskEventDetails.
	MetricsDeletedTables []string `json:"metrics_deleted_tables,omitempty"`

	// Number of metrics_free_sz.
	// Required: true
	MetricsFreeSz *int64 `json:"metrics_free_sz"`

	// Number of metrics_quota.
	// Required: true
	MetricsQuota *int64 `json:"metrics_quota"`
}
