package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxtTier1SegmentManualMode nsxt tier1 segment manual mode
// swagger:model NsxtTier1SegmentManualMode
type NsxtTier1SegmentManualMode struct {

	// Tier1 logical router placement information. Field introduced in 20.1.1. Minimum of 1 items required. Maximum of 128 items allowed.
	Tier1Lrs []*Tier1LogicalRouterInfo `json:"tier1_lrs,omitempty"`
}
