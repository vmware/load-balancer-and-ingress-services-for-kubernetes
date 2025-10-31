// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbServiceConfigAPIResponse gslb service config Api response
// swagger:model GslbServiceConfigApiResponse
type GslbServiceConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*GslbServiceConfig `json:"results,omitempty"`
}
