// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OpenStackAPIVersionCheckFailure open stack Api version check failure
// swagger:model OpenStackApiVersionCheckFailure
type OpenStackAPIVersionCheckFailure struct {

	// Cloud UUID. Field introduced in 20.1.1.
	CcID *string `json:"cc_id,omitempty"`

	// Cloud name. Field introduced in 20.1.1.
	CcName *string `json:"cc_name,omitempty"`

	// Failure reason containing expected API version and actual version. Field introduced in 20.1.1.
	ErrorString *string `json:"error_string,omitempty"`
}
