// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JWTMatch j w t match
// swagger:model JWTMatch
type JWTMatch struct {

	// Claims whose values need to be matched. Field introduced in 20.1.3.
	Matches []*JWTClaimMatch `json:"matches,omitempty"`

	// Token for which the claims need to be validated. Field introduced in 20.1.3.
	TokenName *string `json:"token_name,omitempty"`
}
