package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPRewriteURLAction HTTP rewrite URL action
// swagger:model HTTPRewriteURLAction
type HTTPRewriteURLAction struct {

	// Host config.
	HostHdr *URIParam `json:"host_hdr,omitempty"`

	// Path config.
	Path *URIParam `json:"path,omitempty"`

	// Query config.
	Query *URIParamQuery `json:"query,omitempty"`
}
