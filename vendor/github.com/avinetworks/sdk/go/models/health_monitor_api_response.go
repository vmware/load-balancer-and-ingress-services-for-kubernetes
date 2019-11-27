package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HealthMonitorAPIResponse health monitor Api response
// swagger:model HealthMonitorApiResponse
type HealthMonitorAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*HealthMonitor `json:"results,omitempty"`
}
