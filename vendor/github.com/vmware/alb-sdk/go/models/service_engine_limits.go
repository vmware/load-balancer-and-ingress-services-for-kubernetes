package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServiceEngineLimits service engine limits
// swagger:model ServiceEngineLimits
type ServiceEngineLimits struct {

	// Maximum number of virtualservices per serviceengine, including east-west virtualservices. Field introduced in 20.1.1.
	AllVirtualservicesPerServiceengine *int32 `json:"all_virtualservices_per_serviceengine,omitempty"`

	// Maximum number of east-west virtualservices per serviceengine, excluding north-south virtualservices. Field introduced in 20.1.1.
	EwVirtualservicesPerServiceengine *int32 `json:"ew_virtualservices_per_serviceengine,omitempty"`

	// Maximum number of north-south virtualservices per serviceengine, excluding east-west virtualservices. Field introduced in 20.1.1.
	NsVirtualservicesPerServiceengine *int32 `json:"ns_virtualservices_per_serviceengine,omitempty"`

	// Maximum number of logical interfaces (vlan, bond) per serviceengine. Field introduced in 20.1.1.
	NumLogicalIntfPerSe *int32 `json:"num_logical_intf_per_se,omitempty"`

	// Maximum number of physical interfaces per serviceengine. Field introduced in 20.1.1.
	NumPhyIntfPerSe *int32 `json:"num_phy_intf_per_se,omitempty"`

	// Maximum number of virtualservices with realtime metrics enabled. Field introduced in 20.1.1.
	NumVirtualservicesRtMetrics *int32 `json:"num_virtualservices_rt_metrics,omitempty"`

	// Maximum number of vlan interfaces per physical interface. Field introduced in 20.1.1.
	NumVlanIntfPerPhyIntf *int32 `json:"num_vlan_intf_per_phy_intf,omitempty"`

	// Maximum number of vlan interfaces per serviceengine. Field introduced in 20.1.1.
	NumVlanIntfPerSe *int32 `json:"num_vlan_intf_per_se,omitempty"`

	// Serviceengine system limits specific to cloud type. Field introduced in 20.1.1.
	ServiceengineCloudLimits []*ServiceEngineCloudLimits `json:"serviceengine_cloud_limits,omitempty"`
}
