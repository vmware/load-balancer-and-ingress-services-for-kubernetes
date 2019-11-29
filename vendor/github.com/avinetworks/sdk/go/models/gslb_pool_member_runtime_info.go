package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GslbPoolMemberRuntimeInfo gslb pool member runtime info
// swagger:model GslbPoolMemberRuntimeInfo
type GslbPoolMemberRuntimeInfo struct {

	// Application type of the VS. Enum options - APPLICATION_PROFILE_TYPE_L4, APPLICATION_PROFILE_TYPE_HTTP, APPLICATION_PROFILE_TYPE_SYSLOG, APPLICATION_PROFILE_TYPE_DNS, APPLICATION_PROFILE_TYPE_SSL, APPLICATION_PROFILE_TYPE_SIP. Field introduced in 17.2.2.
	AppType *string `json:"app_type,omitempty"`

	// The Site Controller Cluster UUID to which this member belongs.
	ClusterUUID *string `json:"cluster_uuid,omitempty"`

	// Controller retrieved member status at the site.
	ControllerStatus *OperationalStatus `json:"controller_status,omitempty"`

	// DNS computed member status from different sites.
	DatapathStatus []*GslbPoolMemberDatapathStatus `json:"datapath_status,omitempty"`

	// FQDN address of the member. .
	Fqdn *string `json:"fqdn,omitempty"`

	// gs_name of GslbPoolMemberRuntimeInfo.
	GsName *string `json:"gs_name,omitempty"`

	// The GSLB service to which this member belongs.
	GsUUID *string `json:"gs_uuid,omitempty"`

	// GSLB pool member's configured VIP. .
	IP *IPAddr `json:"ip,omitempty"`

	// This is an internal field that conveys the IP address from the controller to service engine in binary format. .
	IPValueToSe *int32 `json:"ip_value_to_se,omitempty"`

	// Operational VIPs of the member  that can map to multiple VS IP addresses such as private, public and floating addresses.
	OperIps []*IPAddr `json:"oper_ips,omitempty"`

	// Gslb Pool member's consolidated operational status .
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	// services configured on the virtual service.
	Services []*Service `json:"services,omitempty"`

	// The Site 's name is required for event-generation etc.
	SiteName *string `json:"site_name,omitempty"`

	// Site persistence pools associated with the VS. Field introduced in 17.2.2.
	SpPools []*GslbServiceSitePersistencePool `json:"sp_pools,omitempty"`

	// Describes the VIP type  Avi or third-party. Enum options - NON_AVI_VIP, AVI_VIP.
	VipType *string `json:"vip_type,omitempty"`

	// VS name belonging to this GSLB service.
	VsName *string `json:"vs_name,omitempty"`

	// VS UUID belonging to this GSLB service.
	VsUUID *string `json:"vs_uuid,omitempty"`

	// Front end L4 metrics of the virtual service.
	VserverL4Metrics *VserverL4MetricsObj `json:"vserver_l4_metrics,omitempty"`

	// Front end L7 metrics of the virtual service.
	VserverL7Metrics *VserverL7MetricsObj `json:"vserver_l7_metrics,omitempty"`
}
