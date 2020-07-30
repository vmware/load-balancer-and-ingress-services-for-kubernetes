package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerLimits controller limits
// swagger:model ControllerLimits
type ControllerLimits struct {

	// Maximum number of certificates per virtualservice. Field introduced in 20.1.1.
	CertificatesPerVirtualservice *int32 `json:"certificates_per_virtualservice,omitempty"`

	// Controller system limits specific to cloud type for all controller sizes. Field introduced in 20.1.1.
	ControllerCloudLimits []*ControllerCloudLimits `json:"controller_cloud_limits,omitempty"`

	// Controller system limits specific to controller sizing. Field introduced in 20.1.1.
	ControllerSizingLimits []*ControllerSizingLimits `json:"controller_sizing_limits,omitempty"`

	// Maximum number of default routes per vrfcontext. Field introduced in 20.1.1.
	DefaultRoutesPerVrfcontext *int32 `json:"default_routes_per_vrfcontext,omitempty"`

	// Maximum number of IP's per ipaddrgroup. Field introduced in 20.1.1.
	IpsPerIpgroup *int32 `json:"ips_per_ipgroup,omitempty"`

	// Maximum number of poolgroups per virtualservice. Field introduced in 20.1.1.
	PoolgroupsPerVirtualservice *int32 `json:"poolgroups_per_virtualservice,omitempty"`

	// Maximum number of pools per poolgroup. Field introduced in 20.1.1.
	PoolsPerPoolgroup *int32 `json:"pools_per_poolgroup,omitempty"`

	// Maximum number of pools per virtualservice. Field introduced in 20.1.1.
	PoolsPerVirtualservice *int32 `json:"pools_per_virtualservice,omitempty"`

	// Maximum number of routes per vrfcontext. Field introduced in 20.1.1.
	RoutesPerVrfcontext *int32 `json:"routes_per_vrfcontext,omitempty"`

	// Maximum number of rules per httppolicy. Field introduced in 20.1.1.
	RulesPerHttppolicy *int32 `json:"rules_per_httppolicy,omitempty"`

	// Maximum number of rules per networksecuritypolicy. Field introduced in 20.1.1.
	RulesPerNetworksecuritypolicy *int32 `json:"rules_per_networksecuritypolicy,omitempty"`

	// Maximum number of servers per pool. Field introduced in 20.1.1.
	ServersPerPool *int32 `json:"servers_per_pool,omitempty"`

	// Maximum number of SNI children virtualservices per SNI parent virtualservice. Field introduced in 20.1.1.
	SniChildrenPerParent *int32 `json:"sni_children_per_parent,omitempty"`

	// Maximum number of strings per stringgroup. Field introduced in 20.1.1.
	StringsPerStringgroup *int32 `json:"strings_per_stringgroup,omitempty"`

	// Maximum number of serviceengine per virtualservice in bgp scaleout mode. Field introduced in 20.1.1.
	VsBgpScaleout *int32 `json:"vs_bgp_scaleout,omitempty"`

	// Maximum number of serviceengine per virtualservice in layer 2 scaleout mode. Field introduced in 20.1.1.
	VsL2Scaleout *int32 `json:"vs_l2_scaleout,omitempty"`
}
