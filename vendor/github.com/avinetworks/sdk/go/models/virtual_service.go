package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VirtualService virtual service
// swagger:model VirtualService
type VirtualService struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// This configuration only applies if the VirtualService is in Legacy Active Standby HA mode and Load Distribution among Active Standby is enabled. This field is used to tag the VirtualService so that VirtualServices with the same tag will share the same Active ServiceEngine. VirtualServices with different tags will have different Active ServiceEngines. If one of the ServiceEngine's in the ServiceEngineGroup fails, all VirtualServices will end up using the same Active ServiceEngine. Redistribution of the VirtualServices can be either manual or automated when the failed ServiceEngine recovers. Redistribution is based on the auto redistribute property of the ServiceEngineGroup. Enum options - ACTIVE_STANDBY_SE_1, ACTIVE_STANDBY_SE_2.
	ActiveStandbySeTag *string `json:"active_standby_se_tag,omitempty"`

	// Determines analytics settings for the application.
	AnalyticsPolicy *AnalyticsPolicy `json:"analytics_policy,omitempty"`

	// Specifies settings related to analytics. It is a reference to an object of type AnalyticsProfile.
	AnalyticsProfileRef *string `json:"analytics_profile_ref,omitempty"`

	// The name of the Contract/Graph associated with the Virtual Service. Should be in the <Contract name> <Graph name> format. This is applicable only for Service Integration mode with Cisco APIC Controller . Field introduced in 17.2.12,18.1.2.
	ApicContractGraph *string `json:"apic_contract_graph,omitempty"`

	// Enable application layer specific features for the Virtual Service. It is a reference to an object of type ApplicationProfile.
	ApplicationProfileRef *string `json:"application_profile_ref,omitempty"`

	// Auto-allocate floating/elastic IP from the Cloud infrastructure. Field deprecated in 17.1.1.
	AutoAllocateFloatingIP *bool `json:"auto_allocate_floating_ip,omitempty"`

	// Auto-allocate VIP from the provided subnet. Field deprecated in 17.1.1.
	AutoAllocateIP *bool `json:"auto_allocate_ip,omitempty"`

	// Availability-zone to place the Virtual Service. Field deprecated in 17.1.1.
	AvailabilityZone *string `json:"availability_zone,omitempty"`

	// (internal-use) FIP allocated by Avi in the Cloud infrastructure. Field deprecated in 17.1.1.
	AviAllocatedFip *bool `json:"avi_allocated_fip,omitempty"`

	// (internal-use) VIP allocated by Avi in the Cloud infrastructure. Field deprecated in 17.1.1.
	AviAllocatedVip *bool `json:"avi_allocated_vip,omitempty"`

	// (internal-use)Applicable for Azure only. Azure Availability set to which this VS is associated. Internally set by the cloud connector. Field introduced in 17.2.12, 18.1.2.
	// Read Only: true
	AzureAvailabilitySet *string `json:"azure_availability_set,omitempty"`

	// (This is a beta feature). Sync Key-Value cache to the new SEs when VS is scaled out. For ex  SSL sessions are stored using VS's Key-Value cache. When the VS is scaled out, the SSL session information is synced to the new SE, allowing existing SSL sessions to be reused on the new SE. . Field introduced in 17.2.7, 18.1.1.
	BulkSyncKvcache *bool `json:"bulk_sync_kvcache,omitempty"`

	// HTTP authentication configuration for protected resources.
	ClientAuth *HTTPClientAuthenticationParams `json:"client_auth,omitempty"`

	// close client connection on vs config update. Field introduced in 17.2.4.
	CloseClientConnOnConfigUpdate *bool `json:"close_client_conn_on_config_update,omitempty"`

	// Checksum of cloud configuration for VS. Internally set by cloud connector.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP.
	CloudType *string `json:"cloud_type,omitempty"`

	// Rate limit the incoming connections to this virtual service.
	ConnectionsRateLimit *RateProfile `json:"connections_rate_limit,omitempty"`

	// Profile used to match and rewrite strings in request and/or response body.
	ContentRewrite *ContentRewriteProfile `json:"content_rewrite,omitempty"`

	// Creator name.
	CreatedBy *string `json:"created_by,omitempty"`

	// Select the algorithm for QoS fairness.  This determines how multiple Virtual Services sharing the same Service Engines will prioritize traffic over a congested network.
	DelayFairness *bool `json:"delay_fairness,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// (internal-use) Discovered networks providing reachability for client facing Virtual Service IP. This field is deprecated. It is a reference to an object of type Network. Field deprecated in 17.1.1.
	DiscoveredNetworkRef []string `json:"discovered_network_ref,omitempty"`

	// (internal-use) Discovered networks providing reachability for client facing Virtual Service IP. This field is used internally by Avi, not editable by the user. Field deprecated in 17.1.1.
	DiscoveredNetworks []*DiscoveredNetwork `json:"discovered_networks,omitempty"`

	// (internal-use) Discovered subnets providing reachability for client facing Virtual Service IP. This field is deprecated. Field deprecated in 17.1.1.
	DiscoveredSubnet []*IPAddrPrefix `json:"discovered_subnet,omitempty"`

	// Service discovery specific data including fully qualified domain name, type and Time-To-Live of the DNS record. Note that only one of fqdn and dns_info setting is allowed.
	DNSInfo []*DNSInfo `json:"dns_info,omitempty"`

	// DNS Policies applied on the dns traffic of the Virtual Service. Field introduced in 17.1.1.
	DNSPolicies []*DNSPolicies `json:"dns_policies,omitempty"`

	// Force placement on all SE's in service group (Mesos mode only).
	EastWestPlacement *bool `json:"east_west_placement,omitempty"`

	// Response traffic to clients will be sent back to the source MAC address of the connection, rather than statically sent to a default gateway.
	EnableAutogw *bool `json:"enable_autogw,omitempty"`

	// Enable Route Health Injection using the BGP Config in the vrf context.
	EnableRhi *bool `json:"enable_rhi,omitempty"`

	// Enable Route Health Injection for Source NAT'ted floating IP Address using the BGP Config in the vrf context.
	EnableRhiSnat *bool `json:"enable_rhi_snat,omitempty"`

	// Enable or disable the Virtual Service.
	Enabled *bool `json:"enabled,omitempty"`

	// Error Page Profile to be used for this virtualservice.This profile is used to send the custom error page to the client generated by the proxy. It is a reference to an object of type ErrorPageProfile. Field introduced in 17.2.4.
	ErrorPageProfileRef *string `json:"error_page_profile_ref,omitempty"`

	// Floating IP to associate with this Virtual Service. Field deprecated in 17.1.1.
	FloatingIP *IPAddr `json:"floating_ip,omitempty"`

	// If auto_allocate_floating_ip is True and more than one floating-ip subnets exist, then the subnet for the floating IP address allocation. This field is applicable only if the VirtualService belongs to an OpenStack or AWS cloud. In OpenStack or AWS cloud it is required when auto_allocate_floating_ip is selected. Field deprecated in 17.1.1.
	FloatingSubnetUUID *string `json:"floating_subnet_uuid,omitempty"`

	// Criteria for flow distribution among SEs. Enum options - LOAD_AWARE, CONSISTENT_HASH_SOURCE_IP_ADDRESS, CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT.
	FlowDist *string `json:"flow_dist,omitempty"`

	// Criteria for flow labelling. Enum options - NO_LABEL, APPLICATION_LABEL, SERVICE_LABEL.
	FlowLabelType *string `json:"flow_label_type,omitempty"`

	// DNS resolvable, fully qualified domain name of the virtualservice. Only one of 'fqdn' and 'dns_info' configuration is allowed.
	Fqdn *string `json:"fqdn,omitempty"`

	// Translate the host name sent to the servers to this value.  Translate the host name sent from servers back to the value used by the client.
	HostNameXlate *string `json:"host_name_xlate,omitempty"`

	// HTTP Policies applied on the data traffic of the Virtual Service.
	HTTPPolicies []*HTTPPolicies `json:"http_policies,omitempty"`

	// Ignore Pool servers network reachability constraints for Virtual Service placement.
	IgnPoolNetReach *bool `json:"ign_pool_net_reach,omitempty"`

	// IP Address of the Virtual Service. Field deprecated in 17.1.1.
	IPAddress *IPAddr `json:"ip_address,omitempty"`

	// Subnet and/or Network for allocating VirtualService IP by IPAM Provider module. Field deprecated in 17.1.1.
	IPAMNetworkSubnet *IPNetworkSubnet `json:"ipam_network_subnet,omitempty"`

	// L4 Policies applied to the data traffic of the Virtual Service. Field introduced in 17.2.7.
	L4Policies []*L4Policies `json:"l4_policies,omitempty"`

	// Limit potential DoS attackers who exceed max_cps_per_client significantly to a fraction of max_cps_per_client for a while.
	LimitDoser *bool `json:"limit_doser,omitempty"`

	// Maximum connections per second per client IP. Allowed values are 10-1000. Special values are 0- 'unlimited'.
	MaxCpsPerClient *int32 `json:"max_cps_per_client,omitempty"`

	// Microservice representing the virtual service. It is a reference to an object of type MicroService.
	MicroserviceRef *string `json:"microservice_ref,omitempty"`

	// Minimum number of UP pools to mark VS up. Field introduced in 18.2.1, 17.2.12.
	MinPoolsUp *int32 `json:"min_pools_up,omitempty"`

	// Name for the Virtual Service.
	// Required: true
	Name *string `json:"name"`

	// Determines network settings such as protocol, TCP or UDP, and related options for the protocol. It is a reference to an object of type NetworkProfile.
	NetworkProfileRef *string `json:"network_profile_ref,omitempty"`

	// Manually override the network on which the Virtual Service is placed. It is a reference to an object of type Network. Field deprecated in 17.1.1.
	NetworkRef *string `json:"network_ref,omitempty"`

	// Network security policies for the Virtual Service. It is a reference to an object of type NetworkSecurityPolicy.
	NetworkSecurityPolicyRef *string `json:"network_security_policy_ref,omitempty"`

	// A list of NSX Service Groups representing the Clients which can access the Virtual IP of the Virtual Service. Field introduced in 17.1.1.
	NsxSecuritygroup []string `json:"nsx_securitygroup,omitempty"`

	// Optional settings that determine performance limits like max connections or bandwdith etc.
	PerformanceLimits *PerformanceLimits `json:"performance_limits,omitempty"`

	// The pool group is an object that contains pools. It is a reference to an object of type PoolGroup.
	PoolGroupRef *string `json:"pool_group_ref,omitempty"`

	// The pool is an object that contains destination servers and related attributes such as load-balancing and persistence. It is a reference to an object of type Pool.
	PoolRef *string `json:"pool_ref,omitempty"`

	// (internal-use) Network port assigned to the Virtual Service IP address. Field deprecated in 17.1.1.
	PortUUID *string `json:"port_uuid,omitempty"`

	// Remove listening port if VirtualService is down.
	RemoveListeningPortOnVsDown *bool `json:"remove_listening_port_on_vs_down,omitempty"`

	// Rate limit the incoming requests to this virtual service.
	RequestsRateLimit *RateProfile `json:"requests_rate_limit,omitempty"`

	// Disable re-distribution of flows across service engines for a virtual service. Enable if the network itself performs flow hashing with ECMP in environments such as GCP.
	ScaleoutEcmp *bool `json:"scaleout_ecmp,omitempty"`

	// The Service Engine Group to use for this Virtual Service. Moving to a new SE Group is disruptive to existing connections for this VS. It is a reference to an object of type ServiceEngineGroup.
	SeGroupRef *string `json:"se_group_ref,omitempty"`

	// Security policy applied on the traffic of the Virtual Service. This policy is used to perform security actions such as Distributed Denial of Service (DDoS) attack mitigation, etc. It is a reference to an object of type SecurityPolicy. Field introduced in 18.2.1.
	SecurityPolicyRef *string `json:"security_policy_ref,omitempty"`

	// Determines the network settings profile for the server side of TCP proxied connections.  Leave blank to use the same settings as the client to VS side of the connection. It is a reference to an object of type NetworkProfile.
	ServerNetworkProfileRef *string `json:"server_network_profile_ref,omitempty"`

	// Metadata pertaining to the Service provided by this virtual service. In Openshift/Kubernetes environments, egress pod info is stored. Any user input to this field will be overwritten by Avi Vantage.
	ServiceMetadata *string `json:"service_metadata,omitempty"`

	// Select pool based on destination port.
	ServicePoolSelect []*ServicePoolSelector `json:"service_pool_select,omitempty"`

	// List of Services defined for this Virtual Service.
	Services []*Service `json:"services,omitempty"`

	// Sideband configuration to be used for this virtualservice.It can be used for sending traffic to sideband VIPs for external inspection etc.
	SidebandProfile *SidebandProfile `json:"sideband_profile,omitempty"`

	// NAT'ted floating source IP Address(es) for upstream connection to servers.
	SnatIP []*IPAddr `json:"snat_ip,omitempty"`

	// GSLB pools used to manage site-persistence functionality. Each site-persistence pool contains the virtualservices in all the other sites, that is auto-generated by the GSLB manager. This is a read-only field for the user. It is a reference to an object of type Pool. Field introduced in 17.2.2.
	// Read Only: true
	SpPoolRefs []string `json:"sp_pool_refs,omitempty"`

	// Select or create one or two certificates, EC and/or RSA, that will be presented to SSL/TLS terminated connections. It is a reference to an object of type SSLKeyAndCertificate.
	SslKeyAndCertificateRefs []string `json:"ssl_key_and_certificate_refs,omitempty"`

	// Determines the set of SSL versions and ciphers to accept for SSL/TLS terminated connections. It is a reference to an object of type SSLProfile.
	SslProfileRef *string `json:"ssl_profile_ref,omitempty"`

	// Expected number of SSL session cache entries (may be exceeded). Allowed values are 1024-16383.
	SslSessCacheAvgSize *int32 `json:"ssl_sess_cache_avg_size,omitempty"`

	// Client Authentication and Authorization Policy for the virtualservice. Field introduced in 18.2.1.
	SsoPolicy *SSOPolicy `json:"sso_policy,omitempty"`

	// List of static DNS records applied to this Virtual Service. These are static entries and no health monitoring is performed against the IP addresses.
	StaticDNSRecords []*DNSRecord `json:"static_dns_records,omitempty"`

	// Subnet providing reachability for client facing Virtual Service IP. Field deprecated in 17.1.1.
	Subnet *IPAddrPrefix `json:"subnet,omitempty"`

	// It represents subnet for the Virtual Service IP address allocation when auto_allocate_ip is True.It is only applicable in OpenStack or AWS cloud. This field is required if auto_allocate_ip is True. Field deprecated in 17.1.1.
	SubnetUUID *string `json:"subnet_uuid,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Server network or list of servers for cloning traffic. It is a reference to an object of type TrafficCloneProfile. Field introduced in 17.1.1.
	TrafficCloneProfileRef *string `json:"traffic_clone_profile_ref,omitempty"`

	// Knob to enable the Virtual Service traffic on its assigned service engines. This setting is effective only when the enabled flag is set to True. Field introduced in 17.2.8.
	TrafficEnabled *bool `json:"traffic_enabled,omitempty"`

	// Specify if this is a normal Virtual Service, or if it is the parent or child of an SNI-enabled virtual hosted Virtual Service. Enum options - VS_TYPE_NORMAL, VS_TYPE_VH_PARENT, VS_TYPE_VH_CHILD.
	Type *string `json:"type,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Use Bridge IP as VIP on each Host in Mesos deployments.
	UseBridgeIPAsVip *bool `json:"use_bridge_ip_as_vip,omitempty"`

	// Use the Virtual IP as the SNAT IP for health monitoring and sending traffic to the backend servers instead of the Service Engine interface IP. The caveat of enabling this option is that the VirtualService cannot be configued in an Active-Active HA mode. DNS based Multi VIP solution has to be used for HA & Non-disruptive Upgrade purposes. Field introduced in 17.1.9,17.2.3.
	UseVipAsSnat *bool `json:"use_vip_as_snat,omitempty"`

	// UUID of the VirtualService.
	UUID *string `json:"uuid,omitempty"`

	// The exact name requested from the client's SNI-enabled TLS hello domain name field. If this is a match, the parent VS will forward the connection to this child VS.
	VhDomainName []string `json:"vh_domain_name,omitempty"`

	// Specifies the Virtual Service acting as Virtual Hosting (SNI) parent.
	VhParentVsUUID *string `json:"vh_parent_vs_uuid,omitempty"`

	// List of Virtual Service IPs. While creating a 'Shared VS',please use vsvip_ref to point to the shared entities. Field introduced in 17.1.1.
	Vip []*Vip `json:"vip,omitempty"`

	// Virtual Routing Context that the Virtual Service is bound to. This is used to provide the isolation of the set of networks the application is attached to. It is a reference to an object of type VrfContext.
	VrfContextRef *string `json:"vrf_context_ref,omitempty"`

	// Datascripts applied on the data traffic of the Virtual Service.
	VsDatascripts []*VSDataScripts `json:"vs_datascripts,omitempty"`

	// Checksum of cloud configuration for VsVip. Internally set by cloud connector. Field introduced in 17.2.9, 18.1.2.
	VsvipCloudConfigCksum *string `json:"vsvip_cloud_config_cksum,omitempty"`

	// Mostly used during the creation of Shared VS, this field refers to entities that can be shared across Virtual Services. It is a reference to an object of type VsVip. Field introduced in 17.1.1.
	VsvipRef *string `json:"vsvip_ref,omitempty"`

	// WAF policy for the Virtual Service. It is a reference to an object of type WafPolicy. Field introduced in 17.2.1.
	WafPolicyRef *string `json:"waf_policy_ref,omitempty"`

	// The Quality of Service weight to assign to traffic transmitted from this Virtual Service.  A higher weight will prioritize traffic versus other Virtual Services sharing the same Service Engines. Allowed values are 1-128.
	Weight *int32 `json:"weight,omitempty"`
}
