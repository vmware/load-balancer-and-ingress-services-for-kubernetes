// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafPolicy waf policy
// swagger:model WafPolicy
type WafPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Allow Rules to overwrite the policy mode. This must be set if the policy mode is set to enforcement. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowModeDelegation *bool `json:"allow_mode_delegation,omitempty"`

	// A set of rules which describe conditions under which the request will bypass the WAF. This will be processed in the request header phase before any other WAF related code. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Allowlist *WafPolicyAllowlist `json:"allowlist,omitempty"`

	// Application Specific Signatures. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ApplicationSignatures *WafApplicationSignatures `json:"application_signatures,omitempty"`

	// If this flag is set, the system will try to keep the CRS version used in this policy up-to-date. If a newer CRS object is available on this controller, the system will issue the CRS upgrade process for this WAF Policy. It will not update polices if the current CRS version is CRS-VERSION-NOT-APPLICABLE. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AutoUpdateCrs *bool `json:"auto_update_crs,omitempty"`

	// Enable the functionality to bypass WAF for static file extensions. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BypassStaticExtensions *bool `json:"bypass_static_extensions,omitempty"`

	// Configure thresholds for confidence labels. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfidenceOverride *AppLearningConfidenceOverride `json:"confidence_override,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Creator name. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreatedBy *string `json:"created_by,omitempty"`

	// Override attributes for CRS rules. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CrsOverrides []*WafRuleGroupOverrides `json:"crs_overrides,omitempty"`

	//  Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Enable Application Learning for this WAF policy. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableAppLearning *bool `json:"enable_app_learning,omitempty"`

	// Enable Application Learning based rule updates on the WAF Profile. Rules will be programmed in dedicated WAF learning group. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableAutoRuleUpdates *bool `json:"enable_auto_rule_updates,omitempty"`

	// Enable dynamic regex generation for positive security model rules. This is an experimental feature and shouldn't be used in production. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableRegexLearning *bool `json:"enable_regex_learning,omitempty"`

	// WAF Policy failure mode. This can be 'Open' or 'Closed'. Enum options - WAF_FAILURE_MODE_OPEN, WAF_FAILURE_MODE_CLOSED. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FailureMode *string `json:"failure_mode,omitempty"`

	// Geo Location Mapping Database used by this WafPolicy. It is a reference to an object of type GeoDB. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GeoDbRef *string `json:"geo_db_ref,omitempty"`

	// Parameters for tuning Application learning. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LearningParams *AppLearningParams `json:"learning_params,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Minimum confidence label required for auto rule updates. Enum options - CONFIDENCE_VERY_HIGH, CONFIDENCE_HIGH, CONFIDENCE_PROBABLE, CONFIDENCE_LOW, CONFIDENCE_NONE. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinConfidence *string `json:"min_confidence,omitempty"`

	// WAF Policy mode. This can be detection or enforcement. It can be overwritten by rules if allow_mode_delegation is set. Enum options - WAF_MODE_DETECTION_ONLY, WAF_MODE_ENFORCEMENT. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Mode *string `json:"mode"`

	//  Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// WAF Ruleset paranoia  mode. This is used to select Rules based on the paranoia-level tag. Enum options - WAF_PARANOIA_LEVEL_LOW, WAF_PARANOIA_LEVEL_MEDIUM, WAF_PARANOIA_LEVEL_HIGH, WAF_PARANOIA_LEVEL_EXTREME. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ParanoiaLevel *string `json:"paranoia_level,omitempty"`

	// The Positive Security Model. This is used to describe how the request or parts of the request should look like. It is executed in the Request Body Phase of Avi WAF. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PositiveSecurityModel *WafPositiveSecurityModel `json:"positive_security_model,omitempty"`

	// WAF Rules are categorized in to groups based on their characterization. These groups are created by the user and will be enforced after the CRS groups. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PostCrsGroups []*WafRuleGroup `json:"post_crs_groups,omitempty"`

	// WAF Rules are categorized in to groups based on their characterization. These groups are created by the user and will be  enforced before the CRS groups. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreCrsGroups []*WafRuleGroup `json:"pre_crs_groups,omitempty"`

	// The data files and types referred in this WAF policy. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	RequiredDataFiles []*WafPolicyRequiredDataFile `json:"required_data_files,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// While updating CRS, the system will make sure that new rules are added in DETECTION mode. It only has an effect if the Policy is in ENFORCEMENT mode. In this case, the update will set new rules into DETECTION mode by adding crs_overrides for the new rules. If this flag is not set or if the policy mode is DETECTION, rules will be added without new crs_overrides. This option is used for the auto_update_crs workflow as well as for the UI based CRS update workflow. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpdatedCrsRulesInDetectionMode *bool `json:"updated_crs_rules_in_detection_mode,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// WAF core ruleset used for the CRS part of this Policy. It is a reference to an object of type WafCRS. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WafCrsRef *string `json:"waf_crs_ref,omitempty"`

	// WAF Profile for WAF policy. It is a reference to an object of type WafProfile. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	WafProfileRef *string `json:"waf_profile_ref"`
}
