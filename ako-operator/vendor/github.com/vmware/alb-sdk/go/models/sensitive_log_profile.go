// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SensitiveLogProfile sensitive log profile
// swagger:model SensitiveLogProfile
type SensitiveLogProfile struct {

	// Match sensitive header fields in HTTP application log. Field introduced in 17.2.10, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeaderFieldRules []*SensitiveFieldRule `json:"header_field_rules,omitempty"`

	// Match sensitive URI query params in HTTP application log. Query params from the URI are extracted and checked for matching sensitive parameter names. A successful match will mask the parameter values in accordance with this rule action. Field introduced in 20.1.7, 21.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URIQueryFieldRules []*SensitiveFieldRule `json:"uri_query_field_rules,omitempty"`

	// Match sensitive WAF log fields in HTTP application log. Field introduced in 17.2.13, 18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WafFieldRules []*SensitiveFieldRule `json:"waf_field_rules,omitempty"`
}
