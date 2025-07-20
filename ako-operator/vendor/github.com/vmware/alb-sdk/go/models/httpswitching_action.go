// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpswitchingAction httpswitching action
// swagger:model HTTPSwitchingAction
type HttpswitchingAction struct {

	// Content switching action type. Enum options - HTTP_SWITCHING_SELECT_POOL, HTTP_SWITCHING_SELECT_LOCAL, HTTP_SWITCHING_SELECT_POOLGROUP. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- HTTP_SWITCHING_SELECT_POOL,HTTP_SWITCHING_SELECT_LOCAL), Basic, Enterprise with Cloud Services edition.
	// Required: true
	Action *string `json:"action"`

	// File from which to serve local response to the request. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	File *HTTPLocalFile `json:"file,omitempty"`

	// User-defined HTTP status code to use when serving the local response. Use this to provide codes not available in the list within status_code. Allowed values are 200-599. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OtherStatusCode *uint32 `json:"other_status_code,omitempty"`

	// UUID of the pool group to serve the request. It is a reference to an object of type PoolGroup. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	PoolGroupRef *string `json:"pool_group_ref,omitempty"`

	// UUID of the pool of servers to serve the request. It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolRef *string `json:"pool_ref,omitempty"`

	// Specific pool server to select. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Server *PoolServer `json:"server,omitempty"`

	// HTTP status code to use when serving local response. Enum options - HTTP_LOCAL_RESPONSE_STATUS_CODE_200, HTTP_LOCAL_RESPONSE_STATUS_CODE_204, HTTP_LOCAL_RESPONSE_STATUS_CODE_403, HTTP_LOCAL_RESPONSE_STATUS_CODE_404, HTTP_LOCAL_RESPONSE_STATUS_CODE_429, HTTP_LOCAL_RESPONSE_STATUS_CODE_501. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StatusCode *string `json:"status_code,omitempty"`
}
