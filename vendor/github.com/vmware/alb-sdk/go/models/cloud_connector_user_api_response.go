// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudConnectorUserAPIResponse cloud connector user Api response
// swagger:model CloudConnectorUserApiResponse
type CloudConnectorUserAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*CloudConnectorUser `json:"results,omitempty"`
}
