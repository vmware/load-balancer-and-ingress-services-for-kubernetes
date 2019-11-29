package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AppHdr app hdr
// swagger:model AppHdr
type AppHdr struct {

	//  Enum options - SENSITIVE, INSENSITIVE.
	// Required: true
	HdrMatchCase *string `json:"hdr_match_case"`

	// hdr_name of AppHdr.
	// Required: true
	HdrName *string `json:"hdr_name"`

	//  Enum options - BEGINS_WITH, DOES_NOT_BEGIN_WITH, CONTAINS, DOES_NOT_CONTAIN, ENDS_WITH, DOES_NOT_END_WITH, EQUALS, DOES_NOT_EQUAL, REGEX_MATCH, REGEX_DOES_NOT_MATCH.
	// Required: true
	HdrStringOp *string `json:"hdr_string_op"`
}
