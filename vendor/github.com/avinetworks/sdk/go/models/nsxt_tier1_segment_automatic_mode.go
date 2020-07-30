package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxtTier1SegmentAutomaticMode nsxt tier1 segment automatic mode
// swagger:model NsxtTier1SegmentAutomaticMode
type NsxtTier1SegmentAutomaticMode struct {

	// Uber IP subnet for the logical segments created automatically by Avi controller. Field introduced in 20.1.1.
	// Required: true
	NsxtSegmentSubnet *IPAddrPrefix `json:"nsxt_segment_subnet"`

	// The number of SE data vNic's that can be connected to the Avi logical segment. Field introduced in 20.1.1.
	NumSePerSegment *int32 `json:"num_se_per_segment,omitempty"`

	// Tier1 logical router IDs. Field introduced in 20.1.1.
	Tier1LrIds []string `json:"tier1_lr_ids,omitempty"`
}
