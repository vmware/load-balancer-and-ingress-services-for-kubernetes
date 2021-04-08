package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UpgradeStatusInfoAPIResponse upgrade status info Api response
// swagger:model UpgradeStatusInfoApiResponse
type UpgradeStatusInfoAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*UpgradeStatusInfo `json:"results,omitempty"`
}
