package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPResponseRule HTTP response rule
// swagger:model HTTPResponseRule
type HTTPResponseRule struct {

	// Log all HTTP headers upon rule match.
	AllHeaders *bool `json:"all_headers,omitempty"`

	// Enable or disable the rule.
	// Required: true
	Enable *bool `json:"enable"`

	// HTTP header rewrite action.
	HdrAction []*HTTPHdrAction `json:"hdr_action,omitempty"`

	// Index of the rule.
	// Required: true
	Index *int32 `json:"index"`

	// Location header rewrite action.
	LocHdrAction *HTTPRewriteLocHdrAction `json:"loc_hdr_action,omitempty"`

	// Log HTTP request upon rule match.
	Log *bool `json:"log,omitempty"`

	// Add match criteria to the rule.
	Match *ResponseMatchTarget `json:"match,omitempty"`

	// Name of the rule.
	// Required: true
	Name *string `json:"name"`
}
