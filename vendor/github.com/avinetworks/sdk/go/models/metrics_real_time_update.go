package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricsRealTimeUpdate metrics real time update
// swagger:model MetricsRealTimeUpdate
type MetricsRealTimeUpdate struct {

	// Real time metrics collection duration in minutes. 0 for infinite. Special values are 0 - 'infinite'.
	Duration *int32 `json:"duration,omitempty"`

	// Enables real time metrics collection.  When disabled, 6 hour view is the most granular the system will track.
	// Required: true
	Enabled *bool `json:"enabled"`
}
