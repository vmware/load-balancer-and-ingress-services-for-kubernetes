// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VirtualService virtual service
// swagger:model VirtualService
type VirtualService struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// This configuration only applies if the VirtualService is in Legacy Active Standby HA mode and Load Distribution among Active Standby is enabled. This field is used to tag the VirtualService so that VirtualServices with the same tag will share the same Active ServiceEngine. VirtualServices with different tags will have different Active ServiceEngines. If one of the ServiceEngine's in the ServiceEngineGroup fails, all VirtualServices will end up using the same Active ServiceEngine. Redistribution of the VirtualServices can be either manual or automated when the failed ServiceEngine recovers. Redistribution is based on the auto redistribute property of the ServiceEngineGroup. Enum options - ACTIVE_STANDBY_SE_1, ACTIVE_STANDBY_SE_2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ActiveStandbySeTag *string `json:"active_standby_se_tag,omitempty"`

	// Keep advertising Virtual Service via BGP even if it is marked down by health monitor. This setting takes effect for future Virtual Service flaps. To advertise current VSes that are down, please disable and re-enable the Virtual Service. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AdvertiseDownVs *bool `json:"advertise_down_vs,omitempty"`

	// Process request even if invalid client certificate is presented. Datascript APIs need to be used for processing of such requests. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AllowInvalidClientCert *bool `json:"allow_invalid_client_cert,omitempty"`

	// Determines analytics settings for the application. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AnalyticsPolicy *AnalyticsPolicy `json:"analytics_policy,omitempty"`

	// Specifies settings related to analytics. It is a reference to an object of type AnalyticsProfile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AnalyticsProfileRef *string `json:"analytics_profile_ref,omitempty"`

	// Enable application layer specific features for the Virtual Service. It is a reference to an object of type ApplicationProfile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition. Special default for Essentials edition is System-L4-Application.
	ApplicationProfileRef *string `json:"application_profile_ref,omitempty"`

	// (internal-use)Applicable for Azure only. Azure Availability set to which this VS is associated. Internally set by the cloud connector. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	AzureAvailabilitySet *string `json:"azure_availability_set,omitempty"`

	// LOCAL_PREF to be used for VsVip advertised. Applicable only over iBGP. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	BgpLocalPreference *uint32 `json:"bgp_local_preference,omitempty"`

	// Number of times the local AS should be prepended additionally to VsVip. Applicable only over eBGP. Allowed values are 1-10. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	BgpNumAsPathPrepend *uint32 `json:"bgp_num_as_path_prepend,omitempty"`

	// Select BGP peers, using peer label, for VsVip advertisement. Field introduced in 20.1.5. Maximum of 128 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BgpPeerLabels []string `json:"bgp_peer_labels,omitempty"`

	// Bot detection policy for the Virtual Service. It is a reference to an object of type BotDetectionPolicy. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BotPolicyRef *string `json:"bot_policy_ref,omitempty"`

	// (This is a beta feature). Sync Key-Value cache to the new SEs when VS is scaled out. For ex  SSL sessions are stored using VS's Key-Value cache. When the VS is scaled out, the SSL session information is synced to the new SE, allowing existing SSL sessions to be reused on the new SE. . Field introduced in 17.2.7, 18.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	BulkSyncKvcache *bool `json:"bulk_sync_kvcache,omitempty"`

	// close client connection on vs config update. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	CloseClientConnOnConfigUpdate *bool `json:"close_client_conn_on_config_update,omitempty"`

	// Checksum of cloud configuration for VS. Internally set by cloud connector. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	//  It is a reference to an object of type Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	//  Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- CLOUD_NONE,CLOUD_VCENTER), Basic edition(Allowed values- CLOUD_NONE,CLOUD_NSXT), Enterprise with Cloud Services edition.
	CloudType *string `json:"cloud_type,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Rate limit the incoming connections to this virtual service. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	ConnectionsRateLimit *RateProfile `json:"connections_rate_limit,omitempty"`

	// Profile used to match and rewrite strings in request and/or response body. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ContentRewrite *ContentRewriteProfile `json:"content_rewrite,omitempty"`

	// Creator name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CreatedBy *string `json:"created_by,omitempty"`

	// CSRF Protection policy for the Virtual Service. It is a reference to an object of type CSRFPolicy. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CsrfPolicyRef *string `json:"csrf_policy_ref,omitempty"`

	// Select the algorithm for QoS fairness.  This determines how multiple Virtual Services sharing the same Service Engines will prioritize traffic over a congested network. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	DelayFairness *bool `json:"delay_fairness,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Service discovery specific data including fully qualified domain name, type and Time-To-Live of the DNS record. Note that only one of fqdn and dns_info setting is allowed. Maximum of 1000 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSInfo []*DNSInfo `json:"dns_info,omitempty"`

	// DNS Policies applied on the dns traffic of the Virtual Service. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSPolicies []*DNSPolicies `json:"dns_policies,omitempty"`

	// Force placement on all SE's in service group (Mesos mode only). Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	EastWestPlacement *bool `json:"east_west_placement,omitempty"`

	// Response traffic to clients will be sent back to the source MAC address of the connection, rather than statically sent to a default gateway. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition. Special default for Essentials edition is false, Basic edition is false, Enterprise is True.
	EnableAutogw *bool `json:"enable_autogw,omitempty"`

	// Enable Route Health Injection using the BGP Config in the vrf context. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableRhi *bool `json:"enable_rhi,omitempty"`

	// Enable Route Health Injection for Source NAT'ted floating IP Address using the BGP Config in the vrf context. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableRhiSnat *bool `json:"enable_rhi_snat,omitempty"`

	// Enable HTTP sessions for this virtual service. If enabled, a session cookie will be added to HTTP responses and persistent key-value store will be activated. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableSession *bool `json:"enable_session,omitempty"`

	// Enable or disable the Virtual Service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Error Page Profile to be used for this virtualservice.This profile is used to send the custom error page to the client generated by the proxy. It is a reference to an object of type ErrorPageProfile. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ErrorPageProfileRef *string `json:"error_page_profile_ref,omitempty"`

	// Criteria for flow distribution among SEs. Enum options - LOAD_AWARE, CONSISTENT_HASH_SOURCE_IP_ADDRESS, CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- LOAD_AWARE), Basic edition(Allowed values- LOAD_AWARE), Enterprise with Cloud Services edition.
	FlowDist *string `json:"flow_dist,omitempty"`

	// Criteria for flow labelling. Enum options - NO_LABEL, APPLICATION_LABEL, SERVICE_LABEL. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowLabelType *string `json:"flow_label_type,omitempty"`

	// DNS resolvable, fully qualified domain name of the virtualservice. Only one of 'fqdn' and 'dns_info' configuration is allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Fqdn *string `json:"fqdn,omitempty"`

	// Translate the host name sent to the servers to this value.  Translate the host name sent from servers back to the value used by the client. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HostNameXlate *string `json:"host_name_xlate,omitempty"`

	// HTTP Policies applied on the data traffic of the Virtual Service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HTTPPolicies []*HTTPPolicies `json:"http_policies,omitempty"`

	// The config settings for the ICAP server when checking the HTTP request. It is a reference to an object of type IcapProfile. Field introduced in 20.1.1. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IcapRequestProfileRefs []string `json:"icap_request_profile_refs,omitempty"`

	// Ignore Pool servers network reachability constraints for Virtual Service placement. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IgnPoolNetReach *bool `json:"ign_pool_net_reach,omitempty"`

	// Application-specific config for JWT validation. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	JwtConfig *JWTValidationVsConfig `json:"jwt_config,omitempty"`

	// L4 Policies applied to the data traffic of the Virtual Service. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	L4Policies []*L4Policies `json:"l4_policies,omitempty"`

	// Application-specific LDAP config. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LdapVsConfig *LDAPVSConfig `json:"ldap_vs_config,omitempty"`

	// Limit potential DoS attackers who exceed max_cps_per_client significantly to a fraction of max_cps_per_client for a while. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LimitDoser *bool `json:"limit_doser,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Maximum connections per second per client IP. Allowed values are 10-1000. Special values are 0- unlimited. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxCpsPerClient *uint32 `json:"max_cps_per_client,omitempty"`

	// Microservice representing the virtual service. It is a reference to an object of type MicroService. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MicroserviceRef *string `json:"microservice_ref,omitempty"`

	// Minimum number of UP pools to mark VS up. Field introduced in 18.2.1, 17.2.12. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinPoolsUp *uint32 `json:"min_pools_up,omitempty"`

	// Name for the Virtual Service. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Determines network settings such as protocol, TCP or UDP, and related options for the protocol. It is a reference to an object of type NetworkProfile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition. Special default for Essentials edition is System-TCP-Fast-Path.
	NetworkProfileRef *string `json:"network_profile_ref,omitempty"`

	// Network security policies for the Virtual Service. It is a reference to an object of type NetworkSecurityPolicy. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NetworkSecurityPolicyRef *string `json:"network_security_policy_ref,omitempty"`

	// A list of NSX Groups representing the Clients which can access the Virtual IP of the Virtual Service. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NsxSecuritygroup []string `json:"nsx_securitygroup,omitempty"`

	// VirtualService specific OAuth config. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OauthVsConfig *OAuthVSConfig `json:"oauth_vs_config,omitempty"`

	// Optional settings that determine performance limits like max connections or bandwdith etc. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PerformanceLimits *PerformanceLimits `json:"performance_limits,omitempty"`

	// The pool group is an object that contains pools. It is a reference to an object of type PoolGroup. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	PoolGroupRef *string `json:"pool_group_ref,omitempty"`

	// The pool is an object that contains destination servers and related attributes such as load-balancing and persistence. It is a reference to an object of type Pool. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolRef *string `json:"pool_ref,omitempty"`

	// Remove listening port if VirtualService is down. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RemoveListeningPortOnVsDown *bool `json:"remove_listening_port_on_vs_down,omitempty"`

	// Rate limit the incoming requests to this virtual service. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RequestsRateLimit *RateProfile `json:"requests_rate_limit,omitempty"`

	// Revoke the advertisement of Virtual Service via the cloud if it is marked down by health monitor. Supported for NSXT clouds only.This setting takes effect for future Virtual Service flaps. To advertise current VSes that are down, please disable and re-enable the Virtual Service. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RevokeVipRoute *bool `json:"revoke_vip_route,omitempty"`

	// Application-specific SAML config. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SamlSpConfig *SAMLSPConfig `json:"saml_sp_config,omitempty"`

	// Disable re-distribution of flows across service engines for a virtual service. Enable if the network itself performs flow hashing with ECMP in environments such as GCP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ScaleoutEcmp *bool `json:"scaleout_ecmp,omitempty"`

	// The Service Engine Group to use for this Virtual Service. Moving to a new SE Group is disruptive to existing connections for this VS. It is a reference to an object of type ServiceEngineGroup. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGroupRef *string `json:"se_group_ref,omitempty"`

	// Security policy applied on the traffic of the Virtual Service. This policy is used to perform security actions such as Distributed Denial of Service (DDoS) attack mitigation, etc. It is a reference to an object of type SecurityPolicy. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SecurityPolicyRef *string `json:"security_policy_ref,omitempty"`

	// Determines the network settings profile for the server side of TCP proxied connections.  Leave blank to use the same settings as the client to VS side of the connection. It is a reference to an object of type NetworkProfile. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServerNetworkProfileRef *string `json:"server_network_profile_ref,omitempty"`

	// Metadata pertaining to the Service provided by this virtual service. In Openshift/Kubernetes environments, egress pod info is stored. Any user input to this field will be overwritten by Avi Vantage. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceMetadata *string `json:"service_metadata,omitempty"`

	// Select pool based on destination port. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ServicePoolSelect []*ServicePoolSelector `json:"service_pool_select,omitempty"`

	// List of Services defined for this Virtual Service. Maximum of 2048 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Services []*Service `json:"services,omitempty"`

	// Sideband configuration to be used for this virtualservice.It can be used for sending traffic to sideband VIPs for external inspection etc. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SidebandProfile *SidebandProfile `json:"sideband_profile,omitempty"`

	// NAT'ted floating source IP Address(es) for upstream connection to servers. Maximum of 32 items allowed. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	SnatIP []*IPAddr `json:"snat_ip,omitempty"`

	// IPV6 address for SE snat. Field introduced in 30.2.1. Maximum of 32 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SnatIp6Addresses []*IPAddr `json:"snat_ip6_addresses,omitempty"`

	// GSLB pools used to manage site-persistence functionality. Each site-persistence pool contains the virtualservices in all the other sites, that is auto-generated by the GSLB manager. This is a read-only field for the user. It is a reference to an object of type Pool. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Read Only: true
	SpPoolRefs []string `json:"sp_pool_refs,omitempty"`

	// Select or create one or two certificates, EC and/or RSA, that will be presented to SSL/TLS terminated connections. It is a reference to an object of type SSLKeyAndCertificate. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslKeyAndCertificateRefs []string `json:"ssl_key_and_certificate_refs,omitempty"`

	// Determines the set of SSL versions and ciphers to accept for SSL/TLS terminated connections. It is a reference to an object of type SSLProfile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslProfileRef *string `json:"ssl_profile_ref,omitempty"`

	// Select SSL Profile based on client IP address match. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SslProfileSelectors []*SSLProfileSelector `json:"ssl_profile_selectors,omitempty"`

	// Expected number of SSL session cache entries (may be exceeded). Allowed values are 1024-16383. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslSessCacheAvgSize *uint32 `json:"ssl_sess_cache_avg_size,omitempty"`

	// The SSO Policy attached to the virtualservice. It is a reference to an object of type SSOPolicy. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SsoPolicyRef *string `json:"sso_policy_ref,omitempty"`

	// List of static DNS records applied to this Virtual Service. These are static entries and no health monitoring is performed against the IP addresses. Maximum of 1000 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StaticDNSRecords []*DNSRecord `json:"static_dns_records,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Used for testing SE Datastore Upgrade 2.0 functionality. It is a reference to an object of type TestSeDatastoreLevel1. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TestSeDatastoreLevel1Ref *string `json:"test_se_datastore_level_1_ref,omitempty"`

	// Topology Policies applied on the dns traffic of the Virtual Service based onGSLB Topology algorithm. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TopologyPolicies []*DNSPolicies `json:"topology_policies,omitempty"`

	// Server network or list of servers for cloning traffic. It is a reference to an object of type TrafficCloneProfile. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TrafficCloneProfileRef *string `json:"traffic_clone_profile_ref,omitempty"`

	// Knob to enable the Virtual Service traffic on its assigned service engines. This setting is effective only when the enabled flag is set to True. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TrafficEnabled *bool `json:"traffic_enabled,omitempty"`

	// Specify if this is a normal Virtual Service, or if it is the parent or child of an SNI-enabled virtual hosted Virtual Service. Enum options - VS_TYPE_NORMAL, VS_TYPE_VH_PARENT, VS_TYPE_VH_CHILD. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- VS_TYPE_NORMAL), Basic edition(Allowed values- VS_TYPE_NORMAL,VS_TYPE_VH_PARENT), Enterprise with Cloud Services edition.
	Type *string `json:"type,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Use Bridge IP as VIP on each Host in Mesos deployments. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	UseBridgeIPAsVip *bool `json:"use_bridge_ip_as_vip,omitempty"`

	// Use the Virtual IP as the SNAT IP for health monitoring and sending traffic to the backend servers instead of the Service Engine interface IP. The caveat of enabling this option is that the VirtualService cannot be configued in an Active-Active HA mode. DNS based Multi VIP solution has to be used for HA & Non-disruptive Upgrade purposes. Field introduced in 17.1.9,17.2.3. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic, Enterprise with Cloud Services edition.
	UseVipAsSnat *bool `json:"use_vip_as_snat,omitempty"`

	// UUID of the VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// The exact name requested from the client's SNI-enabled TLS hello domain name field. If this is a match, the parent VS will forward the connection to this child VS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VhDomainName []string `json:"vh_domain_name,omitempty"`

	// Match criteria to select this child VS. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VhMatches []*VHMatch `json:"vh_matches,omitempty"`

	// Specifies the Virtual Service acting as Virtual Hosting (SNI) parent. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VhParentVsRef *string `json:"vh_parent_vs_ref,omitempty"`

	// Specify if the Virtual Hosting VS is of type SNI or Enhanced. Enum options - VS_TYPE_VH_SNI, VS_TYPE_VH_ENHANCED. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Basic edition(Allowed values- VS_TYPE_VH_SNI,VS_TYPE_VH_ENHANCED), Enterprise with Cloud Services edition.
	VhType *string `json:"vh_type,omitempty"`

	// List of Virtual Service IPs. While creating a 'Shared VS',please use vsvip_ref to point to the shared entities. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vip []*Vip `json:"vip,omitempty"`

	// Virtual Routing Context that the Virtual Service is bound to. This is used to provide the isolation of the set of networks the application is attached to. It is a reference to an object of type VrfContext. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VrfContextRef *string `json:"vrf_context_ref,omitempty"`

	// Datascripts applied on the data traffic of the Virtual Service. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	VsDatascripts []*VSDataScripts `json:"vs_datascripts,omitempty"`

	// Checksum of cloud configuration for VsVip. Internally set by cloud connector. Field introduced in 17.2.9, 18.1.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	VsvipCloudConfigCksum *string `json:"vsvip_cloud_config_cksum,omitempty"`

	// Mostly used during the creation of Shared VS, this field refers to entities that can be shared across Virtual Services. It is a reference to an object of type VsVip. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsvipRef *string `json:"vsvip_ref,omitempty"`

	// WAF policy for the Virtual Service. It is a reference to an object of type WafPolicy. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	WafPolicyRef *string `json:"waf_policy_ref,omitempty"`

	// The Quality of Service weight to assign to traffic transmitted from this Virtual Service.  A higher weight will prioritize traffic versus other Virtual Services sharing the same Service Engines. Allowed values are 1-128. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1), Basic edition(Allowed values- 1), Enterprise with Cloud Services edition.
	Weight *uint32 `json:"weight,omitempty"`
}
