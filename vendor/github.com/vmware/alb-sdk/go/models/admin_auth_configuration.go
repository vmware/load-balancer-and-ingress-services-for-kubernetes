package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AdminAuthConfiguration admin auth configuration
// swagger:model AdminAuthConfiguration
type AdminAuthConfiguration struct {

	// Allow any user created locally to login with local credentials. Field introduced in 17.1.1.
	AllowLocalUserLogin *bool `json:"allow_local_user_login,omitempty"`

	//  It is a reference to an object of type AuthProfile.
	AuthProfileRef *string `json:"auth_profile_ref,omitempty"`

	// Rules list for tenant or role mapping.
	MappingRules []*AuthMappingRule `json:"mapping_rules,omitempty"`
}
