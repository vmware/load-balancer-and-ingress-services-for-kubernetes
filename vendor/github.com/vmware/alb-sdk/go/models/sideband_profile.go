// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SidebandProfile sideband profile
// swagger:model SidebandProfile
type SidebandProfile struct {

	// IP Address of the sideband server. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IP []*IPAddr `json:"ip,omitempty"`

	// Maximum size of the request body that will be sent on the sideband. Allowed values are 0-16384. Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SidebandMaxRequestBodySize *int32 `json:"sideband_max_request_body_size,omitempty"`
}
