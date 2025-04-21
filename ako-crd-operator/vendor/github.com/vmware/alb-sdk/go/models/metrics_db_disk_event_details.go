// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MetricsDbDiskEventDetails metrics db disk event details
// swagger:model MetricsDbDiskEventDetails
type MetricsDbDiskEventDetails struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsDeletedTables []string `json:"metrics_deleted_tables,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MetricsFreeSz *uint64 `json:"metrics_free_sz"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	MetricsQuota *uint64 `json:"metrics_quota"`
}
