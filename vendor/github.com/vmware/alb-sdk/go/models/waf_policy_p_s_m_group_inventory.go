// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPolicyPSMGroupInventory waf policy p s m group inventory
// swagger:model WafPolicyPSMGroupInventory
type WafPolicyPSMGroupInventory struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Configuration summary of the Waf Policy PSM Group. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Config *WafPolicyPSMGroupConfig `json:"config,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the Waf Policy PSM Group. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Virtual services references. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VirtualservicesRefs []*VsRefs `json:"virtualservices_refs,omitempty"`
}
