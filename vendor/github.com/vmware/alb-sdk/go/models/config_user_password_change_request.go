package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConfigUserPasswordChangeRequest config user password change request
// swagger:model ConfigUserPasswordChangeRequest
type ConfigUserPasswordChangeRequest struct {

	// client ip.
	ClientIP *string `json:"client_ip,omitempty"`

	// Password link is sent or rejected.
	Status *string `json:"status,omitempty"`

	// Matched username of email address.
	User *string `json:"user,omitempty"`

	// Email address of user.
	UserEmail *string `json:"user_email,omitempty"`
}
