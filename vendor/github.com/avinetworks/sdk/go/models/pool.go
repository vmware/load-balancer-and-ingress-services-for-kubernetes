package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Pool pool
// swagger:model Pool
type Pool struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Name of container cloud application that constitutes A pool in a A-B pool configuration, if different from VS app. Field deprecated in 18.1.2.
	APool *string `json:"a_pool,omitempty"`

	// A/B pool configuration. Field deprecated in 18.1.2.
	AbPool *AbPool `json:"ab_pool,omitempty"`

	// Priority of this pool in a A-B pool pair. Internally used. Field deprecated in 18.1.2.
	AbPriority *int32 `json:"ab_priority,omitempty"`

	// Determines analytics settings for the pool. Field introduced in 18.1.5, 18.2.1.
	AnalyticsPolicy *PoolAnalyticsPolicy `json:"analytics_policy,omitempty"`

	// Specifies settings related to analytics. It is a reference to an object of type AnalyticsProfile. Field introduced in 18.1.4,18.2.1.
	AnalyticsProfileRef *string `json:"analytics_profile_ref,omitempty"`

	// Synchronize Cisco APIC EPG members with pool servers.
	ApicEpgName *string `json:"apic_epg_name,omitempty"`

	// Persistence will ensure the same user sticks to the same server for a desired duration of time. It is a reference to an object of type ApplicationPersistenceProfile.
	ApplicationPersistenceProfileRef *string `json:"application_persistence_profile_ref,omitempty"`

	// If configured then Avi will trigger orchestration of pool server creation and deletion. It is only supported for container clouds like Mesos, Opensift, Kubernates, Docker etc. It is a reference to an object of type AutoScaleLaunchConfig.
	AutoscaleLaunchConfigRef *string `json:"autoscale_launch_config_ref,omitempty"`

	// Network Ids for the launch configuration.
	AutoscaleNetworks []string `json:"autoscale_networks,omitempty"`

	// Reference to Server Autoscale Policy. It is a reference to an object of type ServerAutoScalePolicy.
	AutoscalePolicyRef *string `json:"autoscale_policy_ref,omitempty"`

	// Inline estimation of capacity of servers.
	CapacityEstimation *bool `json:"capacity_estimation,omitempty"`

	// The maximum time-to-first-byte of a server. Allowed values are 1-5000. Special values are 0 - 'Automatic'.
	CapacityEstimationTtfbThresh *int32 `json:"capacity_estimation_ttfb_thresh,omitempty"`

	// Checksum of cloud configuration for Pool. Internally set by cloud connector.
	CloudConfigCksum *string `json:"cloud_config_cksum,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Connnection pool properties. Field introduced in 18.2.1.
	ConnPoolProperties *ConnPoolProperties `json:"conn_pool_properties,omitempty"`

	// Duration for which new connections will be gradually ramped up to a server recently brought online.  Useful for LB algorithms that are least connection based. Allowed values are 1-300. Special values are 0 - 'Immediate'.
	ConnectionRampDuration *int32 `json:"connection_ramp_duration,omitempty"`

	// Creator name.
	CreatedBy *string `json:"created_by,omitempty"`

	// Traffic sent to servers will use this destination server port unless overridden by the server's specific port attribute. The SSL checkbox enables Avi to server encryption. Allowed values are 1-65535.
	DefaultServerPort *int32 `json:"default_server_port,omitempty"`

	// Indicates whether existing IPs are disabled(false) or deleted(true) on dns hostname refreshDetail -- On a dns refresh, some IPs set on pool may no longer be returned by the resolver. These IPs are deleted from the pool when this knob is set to true. They are disabled, if the knob is set to false. Field introduced in 18.2.3.
	DeleteServerOnDNSRefresh *bool `json:"delete_server_on_dns_refresh,omitempty"`

	// A description of the pool.
	Description *string `json:"description,omitempty"`

	// Comma separated list of domain names which will be used to verify the common names or subject alternative names presented by server certificates. It is performed only when common name check host_check_enabled is enabled.
	DomainName []string `json:"domain_name,omitempty"`

	// Inherited config from VirtualService.
	EastWest *bool `json:"east_west,omitempty"`

	// Enable or disable the pool.  Disabling will terminate all open connections and pause health monitors.
	Enabled *bool `json:"enabled,omitempty"`

	// Names of external auto-scale groups for pool servers. Currently available only for AWS and Azure. Field introduced in 17.1.2.
	ExternalAutoscaleGroups []string `json:"external_autoscale_groups,omitempty"`

	// Enable an action - Close Connection, HTTP Redirect or Local HTTP Response - when a pool failure happens. By default, a connection will be closed, in case the pool experiences a failure.
	FailAction *FailAction `json:"fail_action,omitempty"`

	// Periodicity of feedback for fewest tasks server selection algorithm. Allowed values are 1-300.
	FewestTasksFeedbackDelay *int32 `json:"fewest_tasks_feedback_delay,omitempty"`

	// Used to gracefully disable a server. Virtual service waits for the specified time before terminating the existing connections  to the servers that are disabled. Allowed values are 1-7200. Special values are 0 - 'Immediate', -1 - 'Infinite'.
	GracefulDisableTimeout *int32 `json:"graceful_disable_timeout,omitempty"`

	// Indicates if the pool is a site-persistence pool. . Field introduced in 17.2.1.
	// Read Only: true
	GslbSpEnabled *bool `json:"gslb_sp_enabled,omitempty"`

	// Verify server health by applying one or more health monitors.  Active monitors generate synthetic traffic from each Service Engine and mark a server up or down based on the response. The Passive monitor listens only to client to server communication. It raises or lowers the ratio of traffic destined to a server based on successful responses. It is a reference to an object of type HealthMonitor.
	HealthMonitorRefs []string `json:"health_monitor_refs,omitempty"`

	// Enable common name check for server certificate. If enabled and no explicit domain name is specified, Avi will use the incoming host header to do the match.
	HostCheckEnabled *bool `json:"host_check_enabled,omitempty"`

	// The Passive monitor will monitor client to server connections and requests and adjust traffic load to servers based on successful responses.  This may alter the expected behavior of the LB method, such as Round Robin.
	InlineHealthMonitor *bool `json:"inline_health_monitor,omitempty"`

	// Use list of servers from Ip Address Group. It is a reference to an object of type IpAddrGroup.
	IpaddrgroupRef *string `json:"ipaddrgroup_ref,omitempty"`

	// The load balancing algorithm will pick a server within the pool's list of available servers. Enum options - LB_ALGORITHM_LEAST_CONNECTIONS, LB_ALGORITHM_ROUND_ROBIN, LB_ALGORITHM_FASTEST_RESPONSE, LB_ALGORITHM_CONSISTENT_HASH, LB_ALGORITHM_LEAST_LOAD, LB_ALGORITHM_FEWEST_SERVERS, LB_ALGORITHM_RANDOM, LB_ALGORITHM_FEWEST_TASKS, LB_ALGORITHM_NEAREST_SERVER, LB_ALGORITHM_CORE_AFFINITY, LB_ALGORITHM_TOPOLOGY.
	LbAlgorithm *string `json:"lb_algorithm,omitempty"`

	// HTTP header name to be used for the hash key.
	LbAlgorithmConsistentHashHdr *string `json:"lb_algorithm_consistent_hash_hdr,omitempty"`

	// Degree of non-affinity for core afffinity based server selection. Allowed values are 1-65535. Field introduced in 17.1.3.
	LbAlgorithmCoreNonaffinity *int32 `json:"lb_algorithm_core_nonaffinity,omitempty"`

	// Criteria used as a key for determining the hash between the client and  server. Enum options - LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS, LB_ALGORITHM_CONSISTENT_HASH_SOURCE_IP_ADDRESS_AND_PORT, LB_ALGORITHM_CONSISTENT_HASH_URI, LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_HEADER, LB_ALGORITHM_CONSISTENT_HASH_CUSTOM_STRING, LB_ALGORITHM_CONSISTENT_HASH_CALLID.
	LbAlgorithmHash *string `json:"lb_algorithm_hash,omitempty"`

	// Allow server lookup by name. Field introduced in 17.1.11,17.2.4.
	LookupServerByName *bool `json:"lookup_server_by_name,omitempty"`

	// The maximum number of concurrent connections allowed to each server within the pool. NOTE  applied value will be no less than the number of service engines that the pool is placed on. If set to 0, no limit is applied.
	MaxConcurrentConnectionsPerServer *int32 `json:"max_concurrent_connections_per_server,omitempty"`

	// Rate Limit connections to each server.
	MaxConnRatePerServer *RateProfile `json:"max_conn_rate_per_server,omitempty"`

	// Minimum number of health monitors in UP state to mark server UP. Field introduced in 18.2.1, 17.2.12.
	MinHealthMonitorsUp *int32 `json:"min_health_monitors_up,omitempty"`

	// Minimum number of servers in UP state for marking the pool UP. Field introduced in 18.2.1, 17.2.12.
	MinServersUp *int32 `json:"min_servers_up,omitempty"`

	// The name of the pool.
	// Required: true
	Name *string `json:"name"`

	// (internal-use) Networks designated as containing servers for this pool.  The servers may be further narrowed down by a filter. This field is used internally by Avi, not editable by the user.
	Networks []*NetworkFilter `json:"networks,omitempty"`

	// A list of NSX Service Groups where the Servers for the Pool are created . Field introduced in 17.1.1.
	NsxSecuritygroup []string `json:"nsx_securitygroup,omitempty"`

	// Avi will validate the SSL certificate present by a server against the selected PKI Profile. It is a reference to an object of type PKIProfile.
	PkiProfileRef *string `json:"pki_profile_ref,omitempty"`

	// Manually select the networks and subnets used to provide reachability to the pool's servers.  Specify the Subnet using the following syntax  10-1-1-0/24. Use static routes in VRF configuration when pool servers are not directly connected butroutable from the service engine.
	PlacementNetworks []*PlacementNetwork `json:"placement_networks,omitempty"`

	// Header name for custom header persistence. Field deprecated in 18.1.2.
	PrstHdrName *string `json:"prst_hdr_name,omitempty"`

	// Minimum number of requests to be queued when pool is full.
	RequestQueueDepth *int32 `json:"request_queue_depth,omitempty"`

	// Enable request queue when pool is full.
	RequestQueueEnabled *bool `json:"request_queue_enabled,omitempty"`

	// Rewrite incoming Host Header to server name of the server to which the request is proxied.  Enabling this feature rewrites Host Header for requests to all servers in the pool.
	RewriteHostHeaderToServerName *bool `json:"rewrite_host_header_to_server_name,omitempty"`

	// If SNI server name is specified, rewrite incoming host header to the SNI server name.
	RewriteHostHeaderToSni *bool `json:"rewrite_host_header_to_sni,omitempty"`

	// Server AutoScale. Not used anymore. Field deprecated in 18.1.2.
	ServerAutoScale *bool `json:"server_auto_scale,omitempty"`

	//  Field deprecated in 18.2.1.
	ServerCount *int32 `json:"server_count,omitempty"`

	// Fully qualified DNS hostname which will be used in the TLS SNI extension in server connections if SNI is enabled. If no value is specified, Avi will use the incoming host header instead.
	ServerName *string `json:"server_name,omitempty"`

	// Server reselect configuration for HTTP requests.
	ServerReselect *HttpserverReselect `json:"server_reselect,omitempty"`

	// Server timeout value specifies the time within which a server connection needs to be established and a request-response exchange completes between AVI and the server. Value of 0 results in using default timeout of 60 minutes. Allowed values are 0-3600000. Field introduced in 18.1.5,18.2.1.
	ServerTimeout *int32 `json:"server_timeout,omitempty"`

	// The pool directs load balanced traffic to this list of destination servers. The servers can be configured by IP address, name, network or via IP Address Group.
	Servers []*Server `json:"servers,omitempty"`

	// Metadata pertaining to the service provided by this Pool. In Openshift/Kubernetes environments, app metadata info is stored. Any user input to this field will be overwritten by Avi Vantage. Field introduced in 17.2.14,18.1.5,18.2.1.
	ServiceMetadata *string `json:"service_metadata,omitempty"`

	// Enable TLS SNI for server connections. If disabled, Avi will not send the SNI extension as part of the handshake.
	SniEnabled *bool `json:"sni_enabled,omitempty"`

	// Service Engines will present a client SSL certificate to the server. It is a reference to an object of type SSLKeyAndCertificate.
	SslKeyAndCertificateRef *string `json:"ssl_key_and_certificate_ref,omitempty"`

	// When enabled, Avi re-encrypts traffic to the backend servers. The specific SSL profile defines which ciphers and SSL versions will be supported. It is a reference to an object of type SSLProfile.
	SslProfileRef *string `json:"ssl_profile_ref,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Do not translate the client's destination port when sending the connection to the server.  The pool or servers specified service port will still be used for health monitoring.
	UseServicePort *bool `json:"use_service_port,omitempty"`

	// UUID of the pool.
	UUID *string `json:"uuid,omitempty"`

	// Virtual Routing Context that the pool is bound to. This is used to provide the isolation of the set of networks the pool is attached to. The pool inherits the Virtual Routing Conext of the Virtual Service, and this field is used only internally, and is set by pb-transform. It is a reference to an object of type VrfContext.
	VrfRef *string `json:"vrf_ref,omitempty"`
}
