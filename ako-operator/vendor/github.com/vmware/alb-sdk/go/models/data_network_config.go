// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DataNetworkConfig data network config
// swagger:model DataNetworkConfig
type DataNetworkConfig struct {

	// Nsxt tier1 segment configuration for Avi Service Engine data path. This should be set only when transport zone is of type OVERLAY. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Tier1SegmentConfig *NsxtTier1SegmentConfig `json:"tier1_segment_config,omitempty"`

	// Data transport zone path for Avi Service Engines. Example- /infra/sites/default/enforcement-points/default/transport-zones/xxx-xxx-xxxx. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	TransportZone *string `json:"transport_zone,omitempty"`

	// Data transport zone type overlay or vlan. Enum options - OVERLAY, VLAN. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	TzType *string `json:"tz_type,omitempty"`

	// Data vlan segments path to use for Avi Service Engines. Example- /infra/segments/vlanls. This should be set only when transport zone is of type VLAN. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VlanSegments []string `json:"vlan_segments,omitempty"`
}
