package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CRSDeploymentFailure c r s deployment failure
// swagger:model CRSDeploymentFailure
type CRSDeploymentFailure struct {

	// List of all CRS updates that failed to install. Field introduced in 20.1.1.
	CrsInfo []*CRSDetails `json:"crs_info,omitempty"`

	// Error message to be conveyed to controller UI. Field introduced in 20.1.1.
	Message *string `json:"message,omitempty"`
}
