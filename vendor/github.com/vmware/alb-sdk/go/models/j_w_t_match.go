package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// JWTMatch j w t match
// swagger:model JWTMatch
type JWTMatch struct {

	// Claims whose values need to be matched. Field introduced in 20.1.3.
	Matches []*JWTClaimMatch `json:"matches,omitempty"`

	// Token for which the claims need to be validated. Field introduced in 20.1.3.
	TokenName *string `json:"token_name,omitempty"`
}
