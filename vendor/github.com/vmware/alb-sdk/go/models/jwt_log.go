// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JwtLog jwt log
// swagger:model JwtLog
type JwtLog struct {

	// Authentication policy rule match. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AuthnRuleMatch *AuthnRuleMatch `json:"authn_rule_match,omitempty"`

	// Authorization policy rule match. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AuthzRuleMatch *AuthzRuleMatch `json:"authz_rule_match,omitempty"`

	// Set to true, if JWT validation is successful. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IsJwtVerified *bool `json:"is_jwt_verified,omitempty"`

	// JWT token payload. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TokenPayload *string `json:"token_payload,omitempty"`
}
