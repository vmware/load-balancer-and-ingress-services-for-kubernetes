package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// MetricsDbQueueFullEventDetails metrics db queue full event details
// swagger:model MetricsDbQueueFullEventDetails
type MetricsDbQueueFullEventDetails struct {

	// Number of high.
	High *int64 `json:"high,omitempty"`

	// Number of instanceport.
	Instanceport *int64 `json:"instanceport,omitempty"`

	// Number of low.
	Low *int64 `json:"low,omitempty"`

	// nodeid of MetricsDbQueueFullEventDetails.
	Nodeid *string `json:"nodeid,omitempty"`

	// period of MetricsDbQueueFullEventDetails.
	Period *string `json:"period,omitempty"`

	// Placeholder for description of property runtime of obj type MetricsDbQueueFullEventDetails field type str  type object
	Runtime *MetricsDbRuntime `json:"runtime,omitempty"`

	// Number of watermark.
	Watermark *int64 `json:"watermark,omitempty"`
}
