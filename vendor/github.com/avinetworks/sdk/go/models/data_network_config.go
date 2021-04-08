package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// DataNetworkConfig data network config
// swagger:model DataNetworkConfig
type DataNetworkConfig struct {

	// Nsxt tier1 segment configuration for Avi Service Engine data path. This should be set only when transport zone is of type OVERLAY. Field introduced in 20.1.5. Allowed in Basic edition, Enterprise edition.
	Tier1SegmentConfig *NsxtTier1SegmentConfig `json:"tier1_segment_config,omitempty"`

	// Data transport zone path for Avi Service Engines. Example- /infra/sites/default/enforcement-points/default/transport-zones/xxx-xxx-xxxx. Field introduced in 20.1.5. Allowed in Basic edition, Enterprise edition.
	// Required: true
	TransportZone *string `json:"transport_zone"`

	// Data transport zone type overlay or vlan. Enum options - OVERLAY, VLAN. Field introduced in 20.1.5. Allowed in Basic edition, Enterprise edition.
	// Required: true
	TzType *string `json:"tz_type"`

	// Data vlan segments path to use for Avi Service Engines. Example- /infra/segments/vlanls. This should be set only when transport zone is of type VLAN. Field introduced in 20.1.5.
	VlanSegments []string `json:"vlan_segments,omitempty"`
}
