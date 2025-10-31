// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerSizingLimits controller sizing limits
// swagger:model ControllerSizingLimits
type ControllerSizingLimits struct {

	// Controller system limits specific to cloud type for this controller sizing. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerSizingCloudLimits []*ControllerSizingCloudLimits `json:"controller_sizing_cloud_limits,omitempty"`

	// Controller flavor (S/M/L) for this sizing limit. Enum options - CONTROLLER_ESSENTIALS, CONTROLLER_SMALL, CONTROLLER_MEDIUM, CONTROLLER_LARGE. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Flavor *string `json:"flavor,omitempty"`

	// Maximum number of clouds. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumClouds *int32 `json:"num_clouds,omitempty"`

	// Maximum number of east-west virtualservices. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumEastWestVirtualservices *int32 `json:"num_east_west_virtualservices,omitempty"`

	// Maximum number of servers. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumServers *int32 `json:"num_servers,omitempty"`

	// Maximum number of serviceengines. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumServiceengines *int32 `json:"num_serviceengines,omitempty"`

	// Maximum number of tenants. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumTenants *int32 `json:"num_tenants,omitempty"`

	// Maximum number of virtualservices. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumVirtualservices *int32 `json:"num_virtualservices,omitempty"`

	// Maximum number of virtualservices with realtime metrics enabled. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumVirtualservicesRtMetrics *int32 `json:"num_virtualservices_rt_metrics,omitempty"`

	// Maximum number of vrfcontexts. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumVrfs *int32 `json:"num_vrfs,omitempty"`

	// Maximum number of virtualservices configured with WAF. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumWafVirtualservices *int32 `json:"num_waf_virtualservices,omitempty"`
}
