package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WebApplicationSignatureServiceStatus web application signature service status
// swagger:model WebApplicationSignatureServiceStatus
type WebApplicationSignatureServiceStatus struct {

	// If the last attempted update failed, this is a more detailed error message. Field introduced in 20.1.3.
	Error *string `json:"error,omitempty"`

	// The time when the Application Signature service last successfull attemped to update this object. It will be not update, if an error occurs during an update attempt. In this case, the errror will be set. Field introduced in 20.1.3.
	LastSuccessfulUpdateCheck *TimeStamp `json:"last_successful_update_check,omitempty"`
}
