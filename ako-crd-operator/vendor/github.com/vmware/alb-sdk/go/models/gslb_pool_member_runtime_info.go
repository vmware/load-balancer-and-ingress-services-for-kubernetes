// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbPoolMemberRuntimeInfo gslb pool member runtime info
// swagger:model GslbPoolMemberRuntimeInfo
type GslbPoolMemberRuntimeInfo struct {

	// Application type of the VS. Enum options - APPLICATION_PROFILE_TYPE_L4, APPLICATION_PROFILE_TYPE_HTTP, APPLICATION_PROFILE_TYPE_SYSLOG, APPLICATION_PROFILE_TYPE_DNS, APPLICATION_PROFILE_TYPE_SSL, APPLICATION_PROFILE_TYPE_SIP. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppType *string `json:"app_type,omitempty"`

	// The Site Controller Cluster UUID to which this member belongs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterUUID *string `json:"cluster_uuid,omitempty"`

	// Controller retrieved member status at the site. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerStatus *OperationalStatus `json:"controller_status,omitempty"`

	// DNS computed member status from different sites. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DatapathStatus []*GslbPoolMemberDatapathStatus `json:"datapath_status,omitempty"`

	// FQDN address of the member. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Fqdn *string `json:"fqdn,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GsName *string `json:"gs_name,omitempty"`

	// The GSLB service to which this member belongs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GsUUID *string `json:"gs_uuid,omitempty"`

	// This field will provide information on origin(site name) of the health monitoring information. Field introduced in 22.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HealthMonitorInfo []string `json:"health_monitor_info,omitempty"`

	// GSLB pool member's configured VIP. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IP *IPAddr `json:"ip,omitempty"`

	// This is an internal field that conveys the IP address from the controller to service engine in binary format. . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPValueToSe uint32 `json:"ip_value_to_se,omitempty"`

	// This is an internal field that conveys the IPV6 address from the controller to service engine in binary format. . Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPV6ValueToSe []int64 `json:"ipv6_value_to_se,omitempty,omitempty"`

	// Operational VIPs of the member  that can map to multiple VS IP addresses such as private, public and floating addresses. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OperIps []*IPAddr `json:"oper_ips,omitempty"`

	// Gslb Pool member's consolidated operational status . Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	// services configured on the virtual service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Services []*Service `json:"services,omitempty"`

	// The Site 's name is required for event-generation etc. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SiteName *string `json:"site_name,omitempty"`

	// Site persistence pools associated with the VS. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SpPools []*GslbServiceSitePersistencePool `json:"sp_pools,omitempty"`

	// Describes the VIP type  Avi or third-party. Enum options - NON_AVI_VIP, AVI_VIP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipType *string `json:"vip_type,omitempty"`

	// VS name belonging to this GSLB service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsName *string `json:"vs_name,omitempty"`

	// VS UUID belonging to this GSLB service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsUUID *string `json:"vs_uuid,omitempty"`

	// Front end L4 metrics of the virtual service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VserverL4Metrics *VserverL4MetricsObj `json:"vserver_l4_metrics,omitempty"`

	// Front end L7 metrics of the virtual service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VserverL7Metrics *VserverL7MetricsObj `json:"vserver_l7_metrics,omitempty"`
}
