package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AnalyticsProfileAPIResponse analytics profile Api response
// swagger:model AnalyticsProfileApiResponse
type AnalyticsProfileAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*AnalyticsProfile `json:"results,omitempty"`
}
