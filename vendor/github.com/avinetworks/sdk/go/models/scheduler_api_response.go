package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SchedulerAPIResponse scheduler Api response
// swagger:model SchedulerApiResponse
type SchedulerAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*Scheduler `json:"results,omitempty"`
}
