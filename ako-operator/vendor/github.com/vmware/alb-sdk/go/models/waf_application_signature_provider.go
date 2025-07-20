// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafApplicationSignatureProvider waf application signature provider
// swagger:model WafApplicationSignatureProvider
type WafApplicationSignatureProvider struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Available application names and the ruleset version, when the rules for an application changed the last time. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	AvailableApplications []*WafApplicationSignatureAppVersion `json:"available_applications,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// If this is set to false, all provided rules are imported when updating this object. If this is set to true, only newer rules are considered for import. Newer rules are rules where the rule id is not in the range of 2,000,000 to 2,080,000 or where the rule has a tag with a CVE from 2013 or newer. All other rules are ignored on rule import. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FilterRulesOnImport *bool `json:"filter_rules_on_import,omitempty"`

	// Name of Application Specific Ruleset Provider. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Version of signatures. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	RulesetVersion *string `json:"ruleset_version,omitempty"`

	// If this object is managed by the Application Signatures update service, this field contain the status of this syncronization. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServiceStatus *WebApplicationSignatureServiceStatus `json:"service_status,omitempty"`

	// The WAF rules. Not visible in the API. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	Signatures []*WafRule `json:"signatures,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
