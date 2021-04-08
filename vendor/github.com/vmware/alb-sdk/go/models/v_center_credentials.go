package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VCenterCredentials v center credentials
// swagger:model VCenterCredentials
type VCenterCredentials struct {

	// Password to talk to VCenter server. Field introduced in 20.1.1.
	Password *string `json:"password,omitempty"`

	// Username to talk to VCenter server. Field introduced in 20.1.1.
	Username *string `json:"username,omitempty"`
}
