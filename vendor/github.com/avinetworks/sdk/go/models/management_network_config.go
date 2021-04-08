package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ManagementNetworkConfig management network config
// swagger:model ManagementNetworkConfig
type ManagementNetworkConfig struct {

	// Management overlay segment to use for Avi Service Engines. This should be set only when transport zone is of type OVERLAY. Field introduced in 20.1.5. Allowed in Basic edition, Enterprise edition.
	OverlaySegment *Tier1LogicalRouterInfo `json:"overlay_segment,omitempty"`

	// Management transport zone path for Avi Service Engines. Example- /infra/sites/default/enforcement-points/default/transport-zones/xxx-xxx-xxxx. Field introduced in 20.1.5. Allowed in Basic edition, Enterprise edition.
	// Required: true
	TransportZone *string `json:"transport_zone"`

	// Management transport zone type overlay or vlan. Enum options - OVERLAY, VLAN. Field introduced in 20.1.5. Allowed in Basic edition, Enterprise edition.
	// Required: true
	TzType *string `json:"tz_type"`

	// Management vlan segment path to use for Avi Service Engines. Example- /infra/segments/vlanls. This should be set only when transport zone is of type VLAN. Field introduced in 20.1.5.
	VlanSegment *string `json:"vlan_segment,omitempty"`
}
