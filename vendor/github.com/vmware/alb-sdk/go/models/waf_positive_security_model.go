// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPositiveSecurityModel waf positive security model
// swagger:model WafPositiveSecurityModel
type WafPositiveSecurityModel struct {

	// These groups should be used to separate different levels of concern. The order of the groups matters, one group may mark parts of the request as valid, so that subsequent groups will not check these parts. It is a reference to an object of type WafPolicyPSMGroup. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GroupRefs []string `json:"group_refs,omitempty"`
}
