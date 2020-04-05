package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPolicy waf policy
// swagger:model WafPolicy
type WafPolicy struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Allow Rules to overwrite the policy mode. This must be set if the policy mode is set to enforcement. Field introduced in 18.1.5, 18.2.1.
	AllowModeDelegation *bool `json:"allow_mode_delegation,omitempty"`

	// Creator name. Field introduced in 17.2.4.
	CreatedBy *string `json:"created_by,omitempty"`

	// WAF Rules are categorized in to groups based on their characterization. These groups are system created with CRS groups. Field introduced in 17.2.1.
	CrsGroups []*WafRuleGroup `json:"crs_groups,omitempty"`

	//  Field introduced in 17.2.1.
	Description *string `json:"description,omitempty"`

	// Enable Application Learning for this WAF policy. Field introduced in 18.2.3.
	EnableAppLearning *bool `json:"enable_app_learning,omitempty"`

	// WAF Policy failure mode. This can be 'Open' or 'Closed'. Enum options - WAF_FAILURE_MODE_OPEN, WAF_FAILURE_MODE_CLOSED. Field introduced in 18.1.2.
	FailureMode *string `json:"failure_mode,omitempty"`

	// Configure parameters for WAF learning. Field deprecated in 18.2.3. Field introduced in 18.1.2.
	Learning *WafLearning `json:"learning,omitempty"`

	// WAF Policy mode. This can be detection or enforcement. It can be overwritten by rules if allow_mode_delegation is set. Enum options - WAF_MODE_DETECTION_ONLY, WAF_MODE_ENFORCEMENT. Field introduced in 17.2.1.
	// Required: true
	Mode *string `json:"mode"`

	//  Field introduced in 17.2.1.
	// Required: true
	Name *string `json:"name"`

	// WAF Ruleset paranoia  mode. This is used to select Rules based on the paranoia-level tag. Enum options - WAF_PARANOIA_LEVEL_LOW, WAF_PARANOIA_LEVEL_MEDIUM, WAF_PARANOIA_LEVEL_HIGH, WAF_PARANOIA_LEVEL_EXTREME. Field introduced in 17.2.1.
	ParanoiaLevel *string `json:"paranoia_level,omitempty"`

	// The Positive Security Model. This is used to describe how the request or parts of the request should look like. It is executed in the Request Body Phase of Avi WAF. Field introduced in 18.2.3.
	PositiveSecurityModel *WafPositiveSecurityModel `json:"positive_security_model,omitempty"`

	// WAF Rules are categorized in to groups based on their characterization. These groups are created by the user and will be enforced after the CRS groups. Field introduced in 17.2.1.
	PostCrsGroups []*WafRuleGroup `json:"post_crs_groups,omitempty"`

	// WAF Rules are categorized in to groups based on their characterization. These groups are created by the user and will be  enforced before the CRS groups. Field introduced in 17.2.1.
	PreCrsGroups []*WafRuleGroup `json:"pre_crs_groups,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 17.2.1.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Field introduced in 17.2.1.
	UUID *string `json:"uuid,omitempty"`

	// WAF core ruleset used for the CRS part of this Policy. It is a reference to an object of type WafCRS. Field introduced in 18.1.1.
	WafCrsRef *string `json:"waf_crs_ref,omitempty"`

	// WAF Profile for WAF policy. It is a reference to an object of type WafProfile. Field introduced in 17.2.1.
	// Required: true
	WafProfileRef *string `json:"waf_profile_ref"`

	// A set of rules which describe conditions under which the request will bypass the WAF. This will be executed in the request header phase before any other WAF related code. Field introduced in 18.2.3.
	Whitelist *WafPolicyWhitelist `json:"whitelist,omitempty"`
}
