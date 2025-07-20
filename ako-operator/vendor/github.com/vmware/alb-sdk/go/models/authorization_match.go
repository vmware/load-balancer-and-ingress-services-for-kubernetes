// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthorizationMatch authorization match
// swagger:model AuthorizationMatch
type AuthorizationMatch struct {

	// Access Token claims to be matched. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AccessToken *JWTMatch `json:"access_token,omitempty"`

	// Attributes whose values need to be matched . Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AttrMatches []*AuthAttributeMatch `json:"attr_matches,omitempty"`

	// Host header value to be matched. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostHdr *HostHdrMatch `json:"host_hdr,omitempty"`

	// HTTP methods to be matched. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Method *MethodMatch `json:"method,omitempty"`

	// Paths/URLs to be matched. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *PathMatch `json:"path,omitempty"`
}
