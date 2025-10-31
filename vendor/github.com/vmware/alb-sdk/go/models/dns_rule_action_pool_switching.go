// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRuleActionPoolSwitching Dns rule action pool switching
// swagger:model DnsRuleActionPoolSwitching
type DNSRuleActionPoolSwitching struct {

	// Reference of the pool group to serve the passthrough DNS query which cannot be served locally. It is a reference to an object of type PoolGroup. Field introduced in 18.1.3, 17.2.12. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolGroupRef *string `json:"pool_group_ref,omitempty"`

	// Reference of the pool to serve the passthrough DNS query which cannot be served locally. It is a reference to an object of type Pool. Field introduced in 18.1.3, 17.2.12. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolRef *string `json:"pool_ref,omitempty"`
}
