package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Tier1LogicalRouterInfo tier1 logical router info
// swagger:model Tier1LogicalRouterInfo
type Tier1LogicalRouterInfo struct {

	// Segment ID. Field introduced in 20.1.1.
	// Required: true
	SegmentID *string `json:"segment_id"`

	// Tier1 logical router ID. Field introduced in 20.1.1.
	// Required: true
	Tier1LrID *string `json:"tier1_lr_id"`
}
