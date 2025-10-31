// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L4RuleAction l4 rule action
// swagger:model L4RuleAction
type L4RuleAction struct {

	// Indicates pool or pool-group selection on rule match. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SelectPool *L4RuleActionSelectPool `json:"select_pool,omitempty"`
}
