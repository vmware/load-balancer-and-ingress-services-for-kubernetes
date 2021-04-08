package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPReputationServiceStatus IP reputation service status
// swagger:model IPReputationServiceStatus
type IPReputationServiceStatus struct {

	// If the last attempted update failed, this is a more detailed error message. Field introduced in 20.1.1.
	Error *string `json:"error,omitempty"`

	// The time when the IP reputation service last successfull attemped to update this object. This is the case when either this file references in this object got updated or when the IP reputation service knows positively that there are no newer versions for these files. It will be not update, if an error occurs during an update attempt. In this case, the errror will be set. Field introduced in 20.1.1.
	LastSuccessfulUpdateCheck *TimeStamp `json:"last_successful_update_check,omitempty"`
}
