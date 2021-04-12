package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// NsxtConfiguration nsxt configuration
// swagger:model NsxtConfiguration
type NsxtConfiguration struct {

	// Automatically create DFW rules for VirtualService in NSX-T Manager. Field introduced in 20.1.1. Allowed in Basic(Allowed values- false) edition, Enterprise edition.
	AutomateDfwRules *bool `json:"automate_dfw_rules,omitempty"`

	// Data network configuration for Avi Service Engines. Field introduced in 20.1.5. Allowed in Basic edition, Enterprise edition.
	// Required: true
	DataNetworkConfig *DataNetworkConfig `json:"data_network_config"`

	// Domain where NSGroup objects belongs to. Field introduced in 20.1.1.
	DomainID *string `json:"domain_id,omitempty"`

	// Enforcement point is where the rules of a policy to apply. Field introduced in 20.1.1.
	EnforcementpointID *string `json:"enforcementpoint_id,omitempty"`

	// Management network configuration for Avi Service Engines. Field introduced in 20.1.5. Allowed in Basic edition, Enterprise edition.
	// Required: true
	ManagementNetworkConfig *ManagementNetworkConfig `json:"management_network_config"`

	// Management network segment to use for Avi Service Engines. Field deprecated in 20.1.5. Field introduced in 20.1.1.
	ManagementSegment *Tier1LogicalRouterInfo `json:"management_segment,omitempty"`

	// Credentials to access NSX-T manager. It is a reference to an object of type CloudConnectorUser. Field introduced in 20.1.1.
	// Required: true
	NsxtCredentialsRef *string `json:"nsxt_credentials_ref"`

	// NSX-T manager hostname or IP address. Field introduced in 20.1.1.
	// Required: true
	NsxtURL *string `json:"nsxt_url"`

	// Site where transport zone belongs to. Field introduced in 20.1.1.
	SiteID *string `json:"site_id,omitempty"`

	// Nsxt tier1 segment configuration for Avi Service Engine data path. Field deprecated in 20.1.5. Field introduced in 20.1.1.
	Tier1SegmentConfig *NsxtTier1SegmentConfig `json:"tier1_segment_config,omitempty"`

	// Network zone where nodes can talk via overlay. Virtual IPs and Service Engines will belong to this zone. Value should be path like /infra/sites/default/enforcement-points/default/transport-zones/xxx-xxx-xxxx. Field deprecated in 20.1.5. Field introduced in 20.1.1.
	TransportZone *string `json:"transport_zone,omitempty"`
}
