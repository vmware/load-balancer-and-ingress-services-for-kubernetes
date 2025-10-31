// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DNSRuleActionGsGroupSelection Dns rule action gs group selection
// swagger:model DnsRuleActionGsGroupSelection
type DNSRuleActionGsGroupSelection struct {

	// GSLB Service group name. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	GroupName *string `json:"group_name"`
}
