package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PortalFeatureOptIn portal feature opt in
// swagger:model PortalFeatureOptIn
type PortalFeatureOptIn struct {

	// Flag to check if the user has opted in for proactive case creation in abnormal scenarios. Field introduced in 20.1.1.
	CaseAutoCreate *bool `json:"case_auto_create,omitempty"`

	// Flag to check if the user has opted in for auto deployment of CRS data on controller. Field introduced in 20.1.1.
	CrsAutoDeploy *bool `json:"crs_auto_deploy,omitempty"`
}
