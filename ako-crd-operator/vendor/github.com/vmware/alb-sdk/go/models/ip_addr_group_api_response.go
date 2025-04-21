// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAddrGroupAPIResponse Ip addr group Api response
// swagger:model IpAddrGroupApiResponse
type IPAddrGroupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*IPAddrGroup `json:"results,omitempty"`
}
