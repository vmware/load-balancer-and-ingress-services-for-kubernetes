package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerSizingLimits controller sizing limits
// swagger:model ControllerSizingLimits
type ControllerSizingLimits struct {

	// Controller system limits specific to cloud type for this controller sizing. Field introduced in 20.1.1.
	ControllerSizingCloudLimits []*ControllerSizingCloudLimits `json:"controller_sizing_cloud_limits,omitempty"`

	// Controller flavor (S/M/L) for this sizing limit. Enum options - CONTROLLER_SMALL, CONTROLLER_MEDIUM, CONTROLLER_LARGE. Field introduced in 20.1.1.
	Flavor *string `json:"flavor,omitempty"`

	// Maximum number of clouds. Field introduced in 20.1.1.
	NumClouds *int32 `json:"num_clouds,omitempty"`

	// Maximum number of east-west virtualservices. Field introduced in 20.1.1.
	NumEastWestVirtualservices *int32 `json:"num_east_west_virtualservices,omitempty"`

	// Maximum number of servers. Field introduced in 20.1.1.
	NumServers *int32 `json:"num_servers,omitempty"`

	// Maximum number of serviceengines. Field introduced in 20.1.1.
	NumServiceengines *int32 `json:"num_serviceengines,omitempty"`

	// Maximum number of tenants. Field introduced in 20.1.1.
	NumTenants *int32 `json:"num_tenants,omitempty"`

	// Maximum number of virtualservices. Field introduced in 20.1.1.
	NumVirtualservices *int32 `json:"num_virtualservices,omitempty"`

	// Maximum number of virtualservices with realtime metrics enabled. Field introduced in 20.1.1.
	NumVirtualservicesRtMetrics *int32 `json:"num_virtualservices_rt_metrics,omitempty"`

	// Maximum number of vrfcontexts. Field introduced in 20.1.1.
	NumVrfs *int32 `json:"num_vrfs,omitempty"`
}
