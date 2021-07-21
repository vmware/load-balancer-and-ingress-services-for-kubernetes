// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSVsSyncInfo DNS vs sync info
// swagger:model DNSVsSyncInfo
type DNSVsSyncInfo struct {

	// error of DNSVsSyncInfo.
	Error *string `json:"error,omitempty"`

	// Number of total_records.
	TotalRecords *int32 `json:"total_records,omitempty"`
}
