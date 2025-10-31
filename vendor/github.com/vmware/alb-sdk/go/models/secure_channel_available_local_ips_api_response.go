// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecureChannelAvailableLocalIpsAPIResponse secure channel available local ips Api response
// swagger:model SecureChannelAvailableLocalIPsApiResponse
type SecureChannelAvailableLocalIpsAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*SecureChannelAvailableLocalIps `json:"results,omitempty"`
}
