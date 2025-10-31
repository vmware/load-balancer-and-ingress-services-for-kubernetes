// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtSIRuleDetails nsxt s i rule details
// swagger:model NsxtSIRuleDetails
type NsxtSIRuleDetails struct {

	// Rule Action. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Action *string `json:"action,omitempty"`

	// Destinatios excluded or not. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Destexclude *bool `json:"destexclude,omitempty"`

	// Destination of redirection rule. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Dests []string `json:"dests,omitempty"`

	// Rule Direction. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Direction *string `json:"direction,omitempty"`

	// Error message. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorString *string `json:"error_string,omitempty"`

	// Pool name. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Pool *string `json:"pool,omitempty"`

	// ServiceEngineGroup name. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Segroup *string `json:"segroup,omitempty"`

	// Services of redirection rule. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Services []string `json:"services,omitempty"`

	// Sources of redirection rule. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Sources []string `json:"sources,omitempty"`
}
