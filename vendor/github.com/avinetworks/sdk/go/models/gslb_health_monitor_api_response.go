package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbHealthMonitorAPIResponse gslb health monitor Api response
// swagger:model GslbHealthMonitorApiResponse
type GslbHealthMonitorAPIResponse struct {

	// count
	// Required: true
	Count int32 `json:"count"`

	// results
	// Required: true
	Results []*GslbHealthMonitor `json:"results,omitempty"`
}
