// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DynamicDNSRecordAPIResponse dynamic Dns record Api response
// swagger:model DynamicDnsRecordApiResponse
type DynamicDNSRecordAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*DynamicDNSRecord `json:"results,omitempty"`
}
