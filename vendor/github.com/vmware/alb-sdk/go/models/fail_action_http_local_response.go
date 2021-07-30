// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FailActionHTTPLocalResponse fail action HTTP local response
// swagger:model FailActionHTTPLocalResponse
type FailActionHTTPLocalResponse struct {

	// Placeholder for description of property file of obj type FailActionHTTPLocalResponse field type str  type object
	File *HTTPLocalFile `json:"file,omitempty"`

	//  Enum options - FAIL_HTTP_STATUS_CODE_200, FAIL_HTTP_STATUS_CODE_503.
	StatusCode *string `json:"status_code,omitempty"`
}
