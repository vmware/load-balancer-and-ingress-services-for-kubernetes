// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceEngineLimits service engine limits
// swagger:model ServiceEngineLimits
type ServiceEngineLimits struct {

	// Maximum number of virtualservices per serviceengine, including east-west virtualservices. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllVirtualservicesPerServiceengine *int32 `json:"all_virtualservices_per_serviceengine,omitempty"`

	// Maximum number of east-west virtualservices per serviceengine, excluding north-south virtualservices. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EwVirtualservicesPerServiceengine *int32 `json:"ew_virtualservices_per_serviceengine,omitempty"`

	// Maximum number of north-south virtualservices per serviceengine, excluding east-west virtualservices. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NsVirtualservicesPerServiceengine *int32 `json:"ns_virtualservices_per_serviceengine,omitempty"`

	// Maximum number of logical interfaces (vlan, bond) per serviceengine. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumLogicalIntfPerSe *int32 `json:"num_logical_intf_per_se,omitempty"`

	// Maximum number of physical interfaces per serviceengine. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumPhyIntfPerSe *int32 `json:"num_phy_intf_per_se,omitempty"`

	// Maximum number of virtualservices with realtime metrics enabled. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumVirtualservicesRtMetrics *int32 `json:"num_virtualservices_rt_metrics,omitempty"`

	// Maximum number of vlan interfaces per physical interface. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumVlanIntfPerPhyIntf *int32 `json:"num_vlan_intf_per_phy_intf,omitempty"`

	// Maximum number of vlan interfaces per serviceengine. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumVlanIntfPerSe *int32 `json:"num_vlan_intf_per_se,omitempty"`

	// Serviceengine system limits specific to cloud type. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceengineCloudLimits []*ServiceEngineCloudLimits `json:"serviceengine_cloud_limits,omitempty"`
}
