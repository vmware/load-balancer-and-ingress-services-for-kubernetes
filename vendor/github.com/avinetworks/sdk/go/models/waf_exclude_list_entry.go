package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafExcludeListEntry waf exclude list entry
// swagger:model WafExcludeListEntry
type WafExcludeListEntry struct {

	// Client IP Subnet to exclude for WAF rules. Field introduced in 17.2.1.
	ClientSubnet *IPAddrPrefix `json:"client_subnet,omitempty"`

	// The match_element can be 'ARGS xxx', 'ARGS_GET xxx', 'ARGS_POST xxx', 'ARGS_NAMES xxx', 'FILES xxx', 'QUERY_STRING', 'REQUEST_BASENAME', 'REQUEST_BODY', 'REQUEST_URI', 'REQUEST_URI_RAW', 'REQUEST_COOKIES xxx', 'REQUEST_HEADERS xxx' or 'RESPONSE_HEADERS xxx'. These match_elements in the HTTP Transaction (if present) will be excluded when executing WAF Rules. Field introduced in 17.2.1.
	MatchElement *string `json:"match_element,omitempty"`

	// Criteria for match_element matching. Field introduced in 19.1.1, 18.2.2.
	MatchElementCriteria *WafExclusionType `json:"match_element_criteria,omitempty"`

	// Criteria for URI matching. Field introduced in 17.2.8.
	URIMatchCriteria *WafExclusionType `json:"uri_match_criteria,omitempty"`

	// URI Path to exclude for WAF rules. Field introduced in 17.2.1.
	URIPath *string `json:"uri_path,omitempty"`
}
