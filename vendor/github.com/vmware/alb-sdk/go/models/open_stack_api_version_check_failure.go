package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OpenStackAPIVersionCheckFailure open stack Api version check failure
// swagger:model OpenStackApiVersionCheckFailure
type OpenStackAPIVersionCheckFailure struct {

	// Cloud UUID. Field introduced in 20.1.1.
	CcID *string `json:"cc_id,omitempty"`

	// Cloud name. Field introduced in 20.1.1.
	CcName *string `json:"cc_name,omitempty"`

	// Failure reason containing expected API version and actual version. Field introduced in 20.1.1.
	ErrorString *string `json:"error_string,omitempty"`
}
