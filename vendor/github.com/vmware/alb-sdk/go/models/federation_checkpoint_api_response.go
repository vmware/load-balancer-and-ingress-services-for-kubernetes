// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FederationCheckpointAPIResponse federation checkpoint Api response
// swagger:model FederationCheckpointApiResponse
type FederationCheckpointAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*FederationCheckpoint `json:"results,omitempty"`
}
