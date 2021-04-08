package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VCASetup v c a setup
// swagger:model VCASetup
type VCASetup struct {

	// cc_id of VCASetup.
	CcID *string `json:"cc_id,omitempty"`

	// error_string of VCASetup.
	ErrorString *string `json:"error_string,omitempty"`

	// instance of VCASetup.
	// Required: true
	Instance *string `json:"instance"`

	//  Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	Privilege *string `json:"privilege,omitempty"`

	// username of VCASetup.
	Username *string `json:"username,omitempty"`
}
