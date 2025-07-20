// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RspContentRewriteRule rsp content rewrite rule
// swagger:model RspContentRewriteRule
type RspContentRewriteRule struct {

	// Enable rewrite rule on response body. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enable *bool `json:"enable,omitempty"`

	// Index of the response rewrite rule. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Index *int32 `json:"index,omitempty"`

	// Name of the response rewrite rule. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// List of search-and-replace *string pairs for the response body. For eg. Strings 'foo' and 'bar', where all searches of 'foo' in the response body will be replaced with 'bar'. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Pairs []*SearchReplacePair `json:"pairs,omitempty"`
}
