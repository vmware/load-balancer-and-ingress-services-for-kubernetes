// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ProactiveSupportDefaults proactive support defaults
// swagger:model ProactiveSupportDefaults
type ProactiveSupportDefaults struct {

	// Opt-in to attach core dump with support case. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AttachCoreDump *bool `json:"attach_core_dump,omitempty"`

	// Opt-in to attach tech support with support case. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition. Special default for Essentials edition is false, Basic edition is false, Enterprise is True.
	AttachTechSupport *bool `json:"attach_tech_support,omitempty"`

	// Case severity to be used for proactive support case creation. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CaseSeverity *string `json:"case_severity,omitempty"`
}
