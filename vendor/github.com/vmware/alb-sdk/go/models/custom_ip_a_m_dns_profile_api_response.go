// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CustomIPAMDNSProfileAPIResponse custom ipam Dns profile Api response
// swagger:model CustomIpamDnsProfileApiResponse
type CustomIPAMDNSProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*CustomIPAMDNSProfile `json:"results,omitempty"`
}
