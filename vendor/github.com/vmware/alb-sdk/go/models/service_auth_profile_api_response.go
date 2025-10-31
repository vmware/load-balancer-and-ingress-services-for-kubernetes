// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceAuthProfileAPIResponse service auth profile Api response
// swagger:model ServiceAuthProfileApiResponse
type ServiceAuthProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ServiceAuthProfile `json:"results,omitempty"`
}
