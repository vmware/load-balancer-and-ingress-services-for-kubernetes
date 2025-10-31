// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecureChannelTokenAPIResponse secure channel token Api response
// swagger:model SecureChannelTokenApiResponse
type SecureChannelTokenAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SecureChannelToken `json:"results,omitempty"`
}
