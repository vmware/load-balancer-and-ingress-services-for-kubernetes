package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxtTier1SegmentConfig nsxt tier1 segment config
// swagger:model NsxtTier1SegmentConfig
type NsxtTier1SegmentConfig struct {

	// Avi controller creates and manages logical segments for a Tier-1 LR. Field introduced in 20.1.1.
	Automatic *NsxtTier1SegmentAutomaticMode `json:"automatic,omitempty"`

	// Avi Admin selects an available logical segment (created by NSX-T admin) associated with a Tier-1 LR. Field introduced in 20.1.1.
	Manual *NsxtTier1SegmentManualMode `json:"manual,omitempty"`

	// Config Mode for selecting the placement logical segments for Avi ServiceEngine data path. Enum options - TIER1_SEGMENT_MANUAL, TIER1_SEGMENT_AUTOMATIC. Field introduced in 20.1.1.
	// Required: true
	SegmentConfigMode *string `json:"segment_config_mode"`
}
