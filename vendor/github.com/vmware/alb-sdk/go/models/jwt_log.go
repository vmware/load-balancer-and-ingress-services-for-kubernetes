package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// JwtLog jwt log
// swagger:model JwtLog
type JwtLog struct {

	// Authentication policy rule match. Field introduced in 20.1.3.
	AuthnRuleMatch *AuthnRuleMatch `json:"authn_rule_match,omitempty"`

	// Authorization policy rule match. Field introduced in 20.1.3.
	AuthzRuleMatch *AuthzRuleMatch `json:"authz_rule_match,omitempty"`

	// Set to true, if JWT validation is successful. Field introduced in 20.1.3.
	IsJwtVerified *bool `json:"is_jwt_verified,omitempty"`

	// JWT token payload. Field introduced in 20.1.3.
	TokenPayload *string `json:"token_payload,omitempty"`
}
