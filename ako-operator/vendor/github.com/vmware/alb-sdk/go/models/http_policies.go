// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPPolicies HTTP policies
// swagger:model HTTPPolicies
type HTTPPolicies struct {

	// UUID of the virtual service HTTP policy collection. It is a reference to an object of type HTTPPolicySet. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	HTTPPolicySetRef *string `json:"http_policy_set_ref"`

	// Index of the virtual service HTTP policy collection. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`
}
