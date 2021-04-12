package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Tier1LogicalRouterInfo tier1 logical router info
// swagger:model Tier1LogicalRouterInfo
type Tier1LogicalRouterInfo struct {

	// Overlay segment path. Example- /infra/segments/Seg-Web-T1-01. Field introduced in 20.1.1.
	// Required: true
	SegmentID *string `json:"segment_id"`

	// Tier1 logical router path. Example- /infra/tier-1s/T1-01. Field introduced in 20.1.1.
	// Required: true
	Tier1LrID *string `json:"tier1_lr_id"`
}
