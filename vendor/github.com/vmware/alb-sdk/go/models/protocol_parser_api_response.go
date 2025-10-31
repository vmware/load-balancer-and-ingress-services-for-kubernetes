// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ProtocolParserAPIResponse protocol parser Api response
// swagger:model ProtocolParserApiResponse
type ProtocolParserAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ProtocolParser `json:"results,omitempty"`
}
