package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudStackSetup cloud stack setup
// swagger:model CloudStackSetup
type CloudStackSetup struct {

	// access_key_id of CloudStackSetup.
	AccessKeyID *string `json:"access_key_id,omitempty"`

	// api_url of CloudStackSetup.
	APIURL *string `json:"api_url,omitempty"`

	// cc_id of CloudStackSetup.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of CloudStackSetup.
	ErrorString *string `json:"error_string,omitempty"`

	//  Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	Privilege *string `json:"privilege,omitempty"`
}
