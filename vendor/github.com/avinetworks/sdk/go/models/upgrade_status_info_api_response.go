package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UpgradeStatusInfoAPIResponse upgrade status info Api response
// swagger:model UpgradeStatusInfoApiResponse
type UpgradeStatusInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*UpgradeStatusInfo `json:"results,omitempty"`
}
