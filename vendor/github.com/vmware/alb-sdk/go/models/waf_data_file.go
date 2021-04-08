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

	// WAF data file type. Enum options - WAF_DATAFILE_PM_FROM_FILE, WAF_DATAFILE_DTD, WAF_DATAFILE_XSD. Field introduced in 20.1.1.
	Type *string `json:"type,omitempty"`
}
