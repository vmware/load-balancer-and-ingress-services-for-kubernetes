// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// WafExcludeListEntry waf exclude list entry
// swagger:model WafExcludeListEntry
type WafExcludeListEntry struct {

	// Client IP Subnet to exclude for WAF rules. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClientSubnet *IPAddrPrefix `json:"client_subnet,omitempty"`

	// Free-text comment about this exclusion. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// The match_element can be 'ARGS xxx', 'ARGS_GET xxx', 'ARGS_POST xxx', 'ARGS_NAMES xxx', 'FILES xxx', 'QUERY_STRING', 'REQUEST_BASENAME', 'REQUEST_BODY', 'REQUEST_URI', 'REQUEST_URI_RAW', 'REQUEST_COOKIES xxx', 'REQUEST_COOKIES_NAMES xxx', 'REQUEST_HEADERS xxx', 'REQUEST_HEADERS_NAMES xxx', 'RESPONSE_HEADERS xxx' or XML xxx. These match_elements in the HTTP Transaction (if present) will be excluded when executing WAF Rules. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchElement *string `json:"match_element,omitempty"`

	// Criteria for match_element matching. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MatchElementCriteria *WafExclusionType `json:"match_element_criteria,omitempty"`

	// Criteria for URI matching. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URIMatchCriteria *WafExclusionType `json:"uri_match_criteria,omitempty"`

	// URI Path to exclude for WAF rules. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URIPath *string `json:"uri_path,omitempty"`
}
