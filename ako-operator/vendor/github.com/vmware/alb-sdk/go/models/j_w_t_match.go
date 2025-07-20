// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// JWTMatch j w t match
// swagger:model JWTMatch
type JWTMatch struct {

	// Claims whose values need to be matched. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Matches []*JWTClaimMatch `json:"matches,omitempty"`

	// Token for which the claims need to be validated. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TokenName *string `json:"token_name,omitempty"`
}
