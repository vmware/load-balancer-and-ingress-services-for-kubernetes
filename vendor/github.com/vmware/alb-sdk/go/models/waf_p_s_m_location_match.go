package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafPSMLocationMatch waf p s m location match
// swagger:model WafPSMLocationMatch
type WafPSMLocationMatch struct {

	// Apply the rules only to requests that match the specified Host header. If this is not set, the host header will not be checked. Field introduced in 18.2.3.
	Host *HostHdrMatch `json:"host,omitempty"`

	// Apply the rules only to requests that have the specified methods. If this is not set, the method will not be checked. Field introduced in 18.2.3.
	Methods *MethodMatch `json:"methods,omitempty"`

	// Apply the rules only to requests that match the specified URI. If this is not set, the path will not be checked. Field introduced in 18.2.3.
	Path *PathMatch `json:"path,omitempty"`
}
