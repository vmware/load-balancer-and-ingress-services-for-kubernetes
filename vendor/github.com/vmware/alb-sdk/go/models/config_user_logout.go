package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConfigUserLogout config user logout
// swagger:model ConfigUserLogout
type ConfigUserLogout struct {

	// client ip.
	ClientIP *string `json:"client_ip,omitempty"`

	// error message if logging out failed.
	ErrorMessage *string `json:"error_message,omitempty"`

	// Local user. Field introduced in 17.1.1.
	Local *bool `json:"local,omitempty"`

	// Status.
	Status *string `json:"status,omitempty"`

	// Request user.
	User *string `json:"user,omitempty"`
}
