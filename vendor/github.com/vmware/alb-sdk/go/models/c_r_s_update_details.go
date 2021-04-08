package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CRSUpdateDetails c r s update details
// swagger:model CRSUpdateDetails
type CRSUpdateDetails struct {

	// List of all available CRS updates. Field introduced in 20.1.1.
	CrsInfo []*CRSDetails `json:"crs_info,omitempty"`
}
