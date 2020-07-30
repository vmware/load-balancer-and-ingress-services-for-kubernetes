package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SSLIgnoredDetails s s l ignored details
// swagger:model SSLIgnoredDetails
type SSLIgnoredDetails struct {

	// Name of SSL Certificate.
	Name *string `json:"name,omitempty"`

	// Reason for ignoring certificate.
	Reason *string `json:"reason,omitempty"`
}
