// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BackupConfigurationAPIResponse backup configuration Api response
// swagger:model BackupConfigurationApiResponse
type BackupConfigurationAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*BackupConfiguration `json:"results,omitempty"`
}
