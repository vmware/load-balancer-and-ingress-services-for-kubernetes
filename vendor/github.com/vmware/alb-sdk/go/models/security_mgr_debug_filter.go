// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SecurityMgrDebugFilter security mgr debug filter
// swagger:model SecurityMgrDebugFilter
type SecurityMgrDebugFilter struct {

	// Dynamically adapt configuration parameters for Application Learning feature. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableAdaptiveConfig *bool `json:"enable_adaptive_config,omitempty"`

	// uuid of the entity. It is a reference to an object of type Virtualservice. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EntityRef *string `json:"entity_ref,omitempty"`

	// Dynamically update the multiplier for rule ID generation in PSM programming for Learning feature. Allowed values are 10-100000. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PsmRuleIDMultiplier *uint32 `json:"psm_rule_id_multiplier,omitempty"`
}
