// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ApplicationPersistenceProfileAPIResponse application persistence profile Api response
// swagger:model ApplicationPersistenceProfileApiResponse
type ApplicationPersistenceProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ApplicationPersistenceProfile `json:"results,omitempty"`
}
