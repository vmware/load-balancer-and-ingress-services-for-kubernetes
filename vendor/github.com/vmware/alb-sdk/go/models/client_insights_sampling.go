package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClientInsightsSampling client insights sampling
// swagger:model ClientInsightsSampling
type ClientInsightsSampling struct {

	// Client IP addresses to check when inserting RUM script.
	ClientIP *IPAddrMatch `json:"client_ip,omitempty"`

	// URL patterns to check when inserting RUM script.
	SampleUris *StringMatch `json:"sample_uris,omitempty"`

	// URL patterns to avoid when inserting RUM script.
	SkipUris *StringMatch `json:"skip_uris,omitempty"`
}
