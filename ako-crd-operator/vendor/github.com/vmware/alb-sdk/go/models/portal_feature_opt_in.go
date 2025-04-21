// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PortalFeatureOptIn portal feature opt in
// swagger:model PortalFeatureOptIn
type PortalFeatureOptIn struct {

	// Enable to receive Application specific signature updates. Field introduced in 20.1.4. Allowed in Enterprise edition only with addon license, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	EnableAppsignatureSync *bool `json:"enable_appsignature_sync,omitempty"`

	// Enable to receive IP reputation updates. Field introduced in 20.1.1. Allowed in Enterprise edition only with addon license, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	EnableIPReputation *bool `json:"enable_ip_reputation,omitempty"`

	// Enable Pulse Cloud Services Case Management. Field introduced in 21.1.1. Allowed in Enterprise edition only with addon license, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition. Special default for Enterprise edition is false, Essentials edition is false, Basic edition is false, Enterprise is True.
	EnablePulseCaseManagement *bool `json:"enable_pulse_case_management,omitempty"`

	// Enable Pulse Cloud Services Inventory. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition. Special default for Essentials edition is false, Basic edition is false, Enterprise is False.
	EnablePulseInventory *bool `json:"enable_pulse_inventory,omitempty"`

	// Enable to receive WAF CRS updates. Field introduced in 21.1.1. Allowed in Enterprise edition only with addon license, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition. Special default for Enterprise edition is false, Essentials edition is false, Basic edition is false, Enterprise is True.
	EnablePulseWafManagement *bool `json:"enable_pulse_waf_management,omitempty"`

	// Enable to receive Bot Management updates. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	EnableUserAgentDbSync *bool `json:"enable_user_agent_db_sync,omitempty"`
}
