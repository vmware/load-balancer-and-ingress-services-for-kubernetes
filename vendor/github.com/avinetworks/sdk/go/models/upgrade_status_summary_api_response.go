package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UpgradeStatusSummaryAPIResponse upgrade status summary Api response
// swagger:model UpgradeStatusSummaryApiResponse
type UpgradeStatusSummaryAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*UpgradeStatusSummary `json:"results,omitempty"`
}
