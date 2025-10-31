// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPolicyCRSUpdate waf policy c r s update
// swagger:model WafPolicyCRSUpdate
type WafPolicyCRSUpdate struct {

	// Set this to true if you want to update the policy. The default value of false will only analyse what would be changed if this flag would be set to true. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Commit *bool `json:"commit,omitempty"`

	// CRS object to which this policy should be updated to. To disable CRS for this policy, the special CRS object CRS-VERSION-NOT-APPLICABLE can be used. It is a reference to an object of type WafCRS. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	WafCrsRef *string `json:"waf_crs_ref"`
}
