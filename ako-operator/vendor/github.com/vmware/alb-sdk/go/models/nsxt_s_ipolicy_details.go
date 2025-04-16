// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtSIpolicyDetails nsxt s ipolicy details
// swagger:model NsxtSIPolicyDetails
type NsxtSIpolicyDetails struct {

	// Error message. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorString *string `json:"error_string,omitempty"`

	// RedirectPolicy Path. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Policy *string `json:"policy,omitempty"`

	// Traffic is redirected to this endpoints. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RedirectTo []string `json:"redirectTo,omitempty"`

	// Policy scope. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Scope *string `json:"scope,omitempty"`

	// ServiceEngineGroup name. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Segroup *string `json:"segroup,omitempty"`

	// Tier1 path. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Tier1 *string `json:"tier1,omitempty"`
}
