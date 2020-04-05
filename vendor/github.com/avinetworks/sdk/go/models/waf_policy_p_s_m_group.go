package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPolicyPSMGroup waf policy p s m group
// swagger:model WafPolicyPSMGroup
type WafPolicyPSMGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Free-text comment about this group. Field introduced in 18.2.3.
	Description *string `json:"description,omitempty"`

	// Enable or disable this WAF rule group. Field introduced in 18.2.3.
	Enable *bool `json:"enable,omitempty"`

	// If a rule in this group matches the match_value pattern, this action will be executed. Allowed actions are WAF_ACTION_NO_OP and WAF_ACTION_ALLOW_PARAMETER. Enum options - WAF_ACTION_NO_OP, WAF_ACTION_BLOCK, WAF_ACTION_ALLOW_PARAMETER. Field introduced in 18.2.3.
	HitAction *string `json:"hit_action,omitempty"`

	// This field indicates that this group is used for learning. Field introduced in 18.2.3.
	IsLearningGroup *bool `json:"is_learning_group,omitempty"`

	// Positive Security Model locations. These are used to partition the application name space. Field introduced in 18.2.3.
	Locations []*WafPSMLocation `json:"locations,omitempty"`

	// If a rule in this group does not match the match_value pattern, this action will be executed. Allowed actions are WAF_ACTION_NO_OP and WAF_ACTION_BLOCK. Enum options - WAF_ACTION_NO_OP, WAF_ACTION_BLOCK, WAF_ACTION_ALLOW_PARAMETER. Field introduced in 18.2.3.
	MissAction *string `json:"miss_action,omitempty"`

	// User defined name of the group. Field introduced in 18.2.3.
	// Required: true
	Name *string `json:"name"`

	// Tenant that this object belongs to. It is a reference to an object of type Tenant. Field introduced in 18.2.3.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of this object. Field introduced in 18.2.3.
	UUID *string `json:"uuid,omitempty"`
}
