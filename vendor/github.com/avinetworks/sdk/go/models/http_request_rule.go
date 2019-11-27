package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPRequestRule HTTP request rule
// swagger:model HTTPRequestRule
type HTTPRequestRule struct {

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

	// Log HTTP request upon rule match.
	Log *bool `json:"log,omitempty"`

	// Add match criteria to the rule.
	Match *MatchTarget `json:"match,omitempty"`

	// Name of the rule.
	// Required: true
	Name *string `json:"name"`

	// HTTP redirect action.
	RedirectAction *HTTPRedirectAction `json:"redirect_action,omitempty"`

	// HTTP request URL rewrite action.
	RewriteURLAction *HTTPRewriteURLAction `json:"rewrite_url_action,omitempty"`

	// Content switching action.
	SwitchingAction *HttpswitchingAction `json:"switching_action,omitempty"`
}
