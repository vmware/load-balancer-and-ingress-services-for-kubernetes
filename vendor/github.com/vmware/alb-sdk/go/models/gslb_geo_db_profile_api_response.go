// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbGeoDbProfileAPIResponse gslb geo db profile Api response
// swagger:model GslbGeoDbProfileApiResponse
type GslbGeoDbProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*GslbGeoDbProfile `json:"results,omitempty"`
}
