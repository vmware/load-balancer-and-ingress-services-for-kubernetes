package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NetworkSecurityRule network security rule
// swagger:model NetworkSecurityRule
type NetworkSecurityRule struct {

	//  Enum options - NETWORK_SECURITY_POLICY_ACTION_TYPE_ALLOW, NETWORK_SECURITY_POLICY_ACTION_TYPE_DENY, NETWORK_SECURITY_POLICY_ACTION_TYPE_RATE_LIMIT.
	// Required: true
	Action *string `json:"action"`

	// Time in minutes after which rule will be deleted. Allowed values are 1-4294967295. Special values are 0- 'blocked for ever'.
	Age *int32 `json:"age,omitempty"`

	// Creator name.
	CreatedBy *string `json:"created_by,omitempty"`

	// Placeholder for description of property enable of obj type NetworkSecurityRule field type str  type boolean
	// Required: true
	Enable *bool `json:"enable"`

	// Number of index.
	// Required: true
	Index *int32 `json:"index"`

	// Placeholder for description of property log of obj type NetworkSecurityRule field type str  type boolean
	Log *bool `json:"log,omitempty"`

	// Placeholder for description of property match of obj type NetworkSecurityRule field type str  type object
	// Required: true
	Match *NetworkSecurityMatchTarget `json:"match"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Placeholder for description of property rl_param of obj type NetworkSecurityRule field type str  type object
	RlParam *NetworkSecurityPolicyActionRLParam `json:"rl_param,omitempty"`
}
