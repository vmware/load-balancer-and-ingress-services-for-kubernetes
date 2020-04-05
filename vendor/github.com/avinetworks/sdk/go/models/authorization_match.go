package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthorizationMatch authorization match
// swagger:model AuthorizationMatch
type AuthorizationMatch struct {

	// Attributes whose values need to be matched . Field introduced in 18.2.5.
	AttrMatches []*AuthAttributeMatch `json:"attr_matches,omitempty"`

	// Host header value to be matched. Field introduced in 18.2.5.
	HostHdr *HostHdrMatch `json:"host_hdr,omitempty"`

	// HTTP methods to be matched. Field introduced in 18.2.5.
	Method *MethodMatch `json:"method,omitempty"`

	// Paths/URLs to be matched. Field introduced in 18.2.5.
	Path *PathMatch `json:"path,omitempty"`
}
