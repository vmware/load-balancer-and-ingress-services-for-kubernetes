// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UpgradeStatusInfoAPIResponse upgrade status info Api response
// swagger:model UpgradeStatusInfoApiResponse
type UpgradeStatusInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*UpgradeStatusInfo `json:"results,omitempty"`
}
