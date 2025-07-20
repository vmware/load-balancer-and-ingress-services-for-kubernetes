// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerLimits controller limits
// swagger:model ControllerLimits
type ControllerLimits struct {

	// BOT system limits. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BotLimits *BOTLimits `json:"bot_limits,omitempty"`

	// Maximum number of certificates per virtualservice. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CertificatesPerVirtualservice *int32 `json:"certificates_per_virtualservice,omitempty"`

	// Controller system limits specific to cloud type for all controller sizes. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerCloudLimits []*ControllerCloudLimits `json:"controller_cloud_limits,omitempty"`

	// Controller system limits specific to controller sizing. Field introduced in 20.1.1. Maximum of 4 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerSizingLimits []*ControllerSizingLimits `json:"controller_sizing_limits,omitempty"`

	// Maximum number of default routes per vrfcontext. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DefaultRoutesPerVrfcontext *int32 `json:"default_routes_per_vrfcontext,omitempty"`

	// Maximum number of gateway monitors per vrfcontext. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GatewayMonPerVrf *int32 `json:"gateway_mon_per_vrf,omitempty"`

	// IP address limits. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IpaddressLimits *IPAddrLimits `json:"ipaddress_limits,omitempty"`

	// Maximum number of IP's per ipaddrgroup. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IpsPerIpgroup *int32 `json:"ips_per_ipgroup,omitempty"`

	// System limits that apply to Layer 7 configuration objects. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	L7Limits *L7limits `json:"l7_limits,omitempty"`

	// Maximum number of poolgroups per virtualservice. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolgroupsPerVirtualservice *int32 `json:"poolgroups_per_virtualservice,omitempty"`

	// Maximum number of pools per poolgroup. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolsPerPoolgroup *int32 `json:"pools_per_poolgroup,omitempty"`

	// Maximum number of pools per virtualservice. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolsPerVirtualservice *int32 `json:"pools_per_virtualservice,omitempty"`

	// Maximum number of routes per vrfcontext. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RoutesPerVrfcontext *int32 `json:"routes_per_vrfcontext,omitempty"`

	// Maximum number of nat rules in nat policy. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RulesPerNatPolicy *int32 `json:"rules_per_nat_policy,omitempty"`

	// Maximum number of rules per networksecuritypolicy. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RulesPerNetworksecuritypolicy *int32 `json:"rules_per_networksecuritypolicy,omitempty"`

	// Maximum number of servers per pool. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServersPerPool *int32 `json:"servers_per_pool,omitempty"`

	// Maximum number of SNI children virtualservices per SNI parent virtualservice. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SniChildrenPerParent *int32 `json:"sni_children_per_parent,omitempty"`

	// Maximum number of strings per stringgroup. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StringsPerStringgroup *int32 `json:"strings_per_stringgroup,omitempty"`

	// Maximum number of serviceengine per virtualservice in bgp scaleout mode. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsBgpScaleout *int32 `json:"vs_bgp_scaleout,omitempty"`

	// Maximum number of serviceengine per virtualservice in layer 2 scaleout mode. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsL2Scaleout *int32 `json:"vs_l2_scaleout,omitempty"`

	// WAF system limits. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	WafLimits *WAFLimits `json:"waf_limits,omitempty"`
}
