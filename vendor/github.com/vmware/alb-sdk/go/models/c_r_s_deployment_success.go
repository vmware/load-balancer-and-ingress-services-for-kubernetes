package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CRSDeploymentSuccess c r s deployment success
// swagger:model CRSDeploymentSuccess
type CRSDeploymentSuccess struct {

	// List of all installed CRS updates. Field introduced in 20.1.1.
	CrsInfo []*CRSDetails `json:"crs_info,omitempty"`
}
