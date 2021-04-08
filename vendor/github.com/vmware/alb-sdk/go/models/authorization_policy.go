package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthorizationPolicy authorization policy
// swagger:model AuthorizationPolicy
type AuthorizationPolicy struct {

	// Authorization Policy Rules. Field introduced in 18.2.5.
	AuthzRules []*AuthorizationRule `json:"authz_rules,omitempty"`
}
