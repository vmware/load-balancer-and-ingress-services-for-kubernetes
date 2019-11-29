package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WafDataFile waf data file
// swagger:model WafDataFile
type WafDataFile struct {

	// Stringified WAF File Data. Field introduced in 17.2.1.
	// Required: true
	Data *string `json:"data"`

	// WAF Data File Name. Field introduced in 17.2.1.
	// Required: true
	Name *string `json:"name"`
}
