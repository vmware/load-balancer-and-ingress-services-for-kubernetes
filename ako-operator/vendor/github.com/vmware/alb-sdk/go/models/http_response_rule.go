// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPResponseRule HTTP response rule
// swagger:model HTTPResponseRule
type HTTPResponseRule struct {

	// Log all HTTP headers upon rule match. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AllHeaders *bool `json:"all_headers,omitempty"`

	// Enable or disable the rule. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Enable *bool `json:"enable"`

	// HTTP header rewrite action. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HdrAction []*HTTPHdrAction `json:"hdr_action,omitempty"`

	// Index of the rule. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`

	// Location header rewrite action. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LocHdrAction *HTTPRewriteLocHdrAction `json:"loc_hdr_action,omitempty"`

	// Log HTTP request upon rule match. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Log *bool `json:"log,omitempty"`

	// Add match criteria to the rule. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Match *ResponseMatchTarget `json:"match,omitempty"`

	// Name of the rule. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`
}
