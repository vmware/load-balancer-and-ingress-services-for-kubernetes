// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthMappingProfileAPIResponse auth mapping profile Api response
// swagger:model AuthMappingProfileApiResponse
type AuthMappingProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*AuthMappingProfile `json:"results,omitempty"`
}
