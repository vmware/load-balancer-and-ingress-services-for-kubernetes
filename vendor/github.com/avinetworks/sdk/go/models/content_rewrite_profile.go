package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ContentRewriteProfile content rewrite profile
// swagger:model ContentRewriteProfile
type ContentRewriteProfile struct {

	// Strings to be matched and replaced with on the request body.
	ReqMatchReplacePair []*MatchReplacePair `json:"req_match_replace_pair,omitempty"`

	// Enable rewrite on request body.
	RequestRewriteEnabled *bool `json:"request_rewrite_enabled,omitempty"`

	// Enable rewrite on response body.
	ResponseRewriteEnabled *bool `json:"response_rewrite_enabled,omitempty"`

	// Rewrite only content types listed in this *string group. Content types not present in this list are not rewritten. It is a reference to an object of type StringGroup.
	RewritableContentRef *string `json:"rewritable_content_ref,omitempty"`

	// Strings to be matched and replaced with on the response body.
	RspMatchReplacePair []*MatchReplacePair `json:"rsp_match_replace_pair,omitempty"`
}
