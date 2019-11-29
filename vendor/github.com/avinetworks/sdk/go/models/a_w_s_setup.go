package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AWSSetup a w s setup
// swagger:model AWSSetup
type AWSSetup struct {

	// access_key_id of AWSSetup.
	AccessKeyID *string `json:"access_key_id,omitempty"`

	// cc_id of AWSSetup.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of AWSSetup.
	ErrorString *string `json:"error_string,omitempty"`

	//  Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	Privilege *string `json:"privilege,omitempty"`

	// region of AWSSetup.
	// Required: true
	Region *string `json:"region"`

	//  Field introduced in 17.1.3.
	VpcID *string `json:"vpc_id,omitempty"`
}
