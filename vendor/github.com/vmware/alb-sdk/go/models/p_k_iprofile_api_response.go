// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PKIprofileAPIResponse p k iprofile Api response
// swagger:model PKIProfileApiResponse
type PKIprofileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*PKIprofile `json:"results,omitempty"`
}
