package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ServiceEngineGroup service engine group
// swagger:model ServiceEngineGroup
type ServiceEngineGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Enable accelerated networking option for Azure SE. Accelerated networking enables single root I/O virtualization (SR-IOV) to a SE VM. This improves networking performance. Field introduced in 17.2.14,18.1.5,18.2.1.
	AcceleratedNetworking *bool `json:"accelerated_networking,omitempty"`

	// Service Engines in active/standby mode for HA failover.
	ActiveStandby *bool `json:"active_standby,omitempty"`

	// Indicates the percent of config memory used for config updates. Allowed values are 0-90. Field deprecated in 18.1.2. Field introduced in 18.1.1. Unit is PERCENT.
	AdditionalConfigMemory *int32 `json:"additional_config_memory,omitempty"`

	// Advertise reach-ability of backend server networks via ADC through BGP for default gateway feature. Field deprecated in 18.2.5.
	AdvertiseBackendNetworks *bool `json:"advertise_backend_networks,omitempty"`

	// Enable aggressive failover configuration for ha.
	AggressiveFailureDetection *bool `json:"aggressive_failure_detection,omitempty"`

	// In compact placement, Virtual Services are placed on existing SEs until max_vs_per_se limit is reached. Enum options - PLACEMENT_ALGO_PACKED, PLACEMENT_ALGO_DISTRIBUTED.
	Algo *string `json:"algo,omitempty"`

	// Allow SEs to be created using burst license. Field introduced in 17.2.5.
	AllowBurst *bool `json:"allow_burst,omitempty"`

	// A percent value of total SE memory reserved for applicationcaching. This is an SE bootup property and requires SE restart.Requires SE Reboot. Allowed values are 0 - 100. Special values are 0- 'disable'. Field introduced in 18.2.3. Unit is PERCENT.
	AppCachePercent *int32 `json:"app_cache_percent,omitempty"`

	// The max memory that can be allocated for the app cache. This value will act as an upper bound on the cache size specified in app_cache_percent. Special values are 0- 'disable'. Field introduced in 20.1.1. Unit is GB.
	AppCacheThreshold *int32 `json:"app_cache_threshold,omitempty"`

	// A percent value of total SE memory reserved for Application learning. This is an SE bootup property and requires SE restart. Allowed values are 0 - 10. Field introduced in 18.2.3. Unit is PERCENT.
	AppLearningMemoryPercent *int32 `json:"app_learning_memory_percent,omitempty"`

	// Amount of SE memory in GB until which shared memory is collected in core archive. Field introduced in 17.1.3. Unit is GB.
	ArchiveShmLimit *int32 `json:"archive_shm_limit,omitempty"`

	// SSL handshakes will be handled by dedicated SSL Threads.Requires SE Reboot.
	AsyncSsl *bool `json:"async_ssl,omitempty"`

	// Number of Async SSL threads per se_dp.Requires SE Reboot. Allowed values are 1-16.
	AsyncSslThreads *int32 `json:"async_ssl_threads,omitempty"`

	// If set, Virtual Services will be automatically migrated when load on an SE is less than minimum or more than maximum thresholds. Only Alerts are generated when the auto_rebalance is not set.
	AutoRebalance *bool `json:"auto_rebalance,omitempty"`

	// Capacities of SE for auto rebalance for each criteria. Field introduced in 17.2.4.
	AutoRebalanceCapacityPerSe []int64 `json:"auto_rebalance_capacity_per_se,omitempty,omitempty"`

	// Set of criteria for SE Auto Rebalance. Enum options - SE_AUTO_REBALANCE_CPU, SE_AUTO_REBALANCE_PPS, SE_AUTO_REBALANCE_MBPS, SE_AUTO_REBALANCE_OPEN_CONNS, SE_AUTO_REBALANCE_CPS. Field introduced in 17.2.3.
	AutoRebalanceCriteria []string `json:"auto_rebalance_criteria,omitempty"`

	// Frequency of rebalance, if 'Auto rebalance' is enabled. Unit is SEC.
	AutoRebalanceInterval *int32 `json:"auto_rebalance_interval,omitempty"`

	// Redistribution of virtual services from the takeover SE to the replacement SE can cause momentary traffic loss. If the auto-redistribute load option is left in its default off state, any desired rebalancing requires calls to REST API.
	AutoRedistributeActiveStandbyLoad *bool `json:"auto_redistribute_active_standby_load,omitempty"`

	// Availability zones for Virtual Service High Availability. It is a reference to an object of type AvailabilityZone. Field introduced in 20.1.1.
	AvailabilityZoneRefs []string `json:"availability_zone_refs,omitempty"`

	// BGP peer state update interval. Allowed values are 5-100. Field introduced in 17.2.14,18.1.5,18.2.1. Unit is SEC.
	BgpStateUpdateInterval *int32 `json:"bgp_state_update_interval,omitempty"`

	// Excess Service Engine capacity provisioned for HA failover.
	BufferSe *int32 `json:"buffer_se,omitempty"`

	//  It is a reference to an object of type Cloud.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Compress IP rules into a single subnet based IP rule for each north-south IPAM subnet configured in PCAP mode in OpenShift/Kubernetes node. Field introduced in 18.2.9, 20.1.1.
	CompressIPRulesForEachNsSubnet *bool `json:"compress_ip_rules_for_each_ns_subnet,omitempty"`

	// Enable config debugs on all cores of SE. Field introduced in 17.2.13,18.1.5,18.2.1.
	ConfigDebugsOnAllCores *bool `json:"config_debugs_on_all_cores,omitempty"`

	// Percentage of memory for connection state. This will come at the expense of memory used for HTTP in-memory cache. Allowed values are 10-90. Unit is PERCENT.
	ConnectionMemoryPercentage *int32 `json:"connection_memory_percentage,omitempty"`

	// Include shared memory for app cache in core file.Requires SE Reboot. Field introduced in 18.2.8, 20.1.1.
	CoreShmAppCache *bool `json:"core_shm_app_cache,omitempty"`

	// Include shared memory for app learning in core file.Requires SE Reboot. Field introduced in 18.2.8, 20.1.1.
	CoreShmAppLearning *bool `json:"core_shm_app_learning,omitempty"`

	// Placeholder for description of property cpu_reserve of obj type ServiceEngineGroup field type str  type boolean
	CPUReserve *bool `json:"cpu_reserve,omitempty"`

	// Allocate all the CPU cores for the Service Engine Virtual Machines  on the same CPU socket. Applicable only for vCenter Cloud.
	CPUSocketAffinity *bool `json:"cpu_socket_affinity,omitempty"`

	// Custom Security Groups to be associated with data vNics for SE instances in OpenStack and AWS Clouds. Field introduced in 17.1.3.
	CustomSecuritygroupsData []string `json:"custom_securitygroups_data,omitempty"`

	// Custom Security Groups to be associated with management vNic for SE instances in OpenStack and AWS Clouds. Field introduced in 17.1.3.
	CustomSecuritygroupsMgmt []string `json:"custom_securitygroups_mgmt,omitempty"`

	// Custom tag will be used to create the tags for SE instance in AWS. Note this is not the same as the prefix for SE name.
	CustomTag []*CustomTag `json:"custom_tag,omitempty"`

	// Subnet used to spin up the data nic for Service Engines, used only for Azure cloud. Overrides the cloud level setting for Service Engine subnet. Field introduced in 18.2.3.
	DataNetworkID *string `json:"data_network_id,omitempty"`

	// Number of instructions before datascript times out. Allowed values are 0-100000000. Field introduced in 18.2.3.
	DatascriptTimeout *int64 `json:"datascript_timeout,omitempty"`

	// Dedicate the core that handles packet receive/transmit from the network to just the dispatching function. Don't use it for TCP/IP and SSL functions.
	DedicatedDispatcherCore *bool `json:"dedicated_dispatcher_core,omitempty"`

	// User defined description for the object.
	Description *string `json:"description,omitempty"`

	// By default, Avi creates and manages security groups along with custom sg provided by user. Set this to True to disallow Avi to create and manage new security groups. Avi will only make use of custom security groups provided by user. This option is supported for AWS and OpenStack cloud types. Field introduced in 17.2.13,18.1.4,18.2.1.
	DisableAviSecuritygroups *bool `json:"disable_avi_securitygroups,omitempty"`

	// Stop using TCP/UDP and IP checksum offload features of NICs. Field introduced in 17.1.14, 17.2.5, 18.1.1.
	DisableCsumOffloads *bool `json:"disable_csum_offloads,omitempty"`

	// Disable Generic Receive Offload (GRO) in DPDK poll-mode driver packet receive path.  GRO is on by default on NICs that do not support LRO (Large Receive Offload) or do not gain performance boost from LRO. Field introduced in 17.2.5, 18.1.1.
	DisableGro *bool `json:"disable_gro,omitempty"`

	// If set, disable the config memory check done in service engine. Field introduced in 18.1.2.
	DisableSeMemoryCheck *bool `json:"disable_se_memory_check,omitempty"`

	// Disable TCP Segmentation Offload (TSO) in DPDK poll-mode driver packet transmit path. TSO is on by default on NICs that support it. Field introduced in 17.2.5, 18.1.1.
	DisableTso *bool `json:"disable_tso,omitempty"`

	// Amount of disk space for each of the Service Engine virtual machines. Unit is GB.
	DiskPerSe *int32 `json:"disk_per_se,omitempty"`

	// Use both the active and standby Service Engines for Virtual Service placement in the legacy active standby HA mode.
	DistributeLoadActiveStandby *bool `json:"distribute_load_active_standby,omitempty"`

	// Distributes queue ownership among cores so multiple cores handle dispatcher duties. Requires SE Reboot. Deprecated from 18.2.8, instead use max_queues_per_vnic. Field introduced in 17.2.8.
	DistributeQueues *bool `json:"distribute_queues,omitempty"`

	// Distributes vnic ownership among cores so multiple cores handle dispatcher duties.Requires SE Reboot. Field introduced in 18.2.5.
	DistributeVnics *bool `json:"distribute_vnics,omitempty"`

	// Enable GratArp for VIP_IP. Field introduced in 18.2.3.
	EnableGratarpPermanent *bool `json:"enable_gratarp_permanent,omitempty"`

	// (This is a beta feature). Enable HSM key priming. If enabled, key handles on the hsm will be synced to SE before processing client connections. Field introduced in 17.2.7, 18.1.1.
	EnableHsmPriming *bool `json:"enable_hsm_priming,omitempty"`

	// Applicable only for Azure cloud with Basic SKU LB. If set, additional Azure LBs will be automatically created if resources in existing LB are exhausted. Field introduced in 17.2.10, 18.1.2.
	EnableMultiLb *bool `json:"enable_multi_lb,omitempty"`

	// Enable TX ring support in pcap mode of operation. TSO feature is not supported with TX Ring enabled. Deprecated from 18.2.8, instead use pcap_tx_mode. Requires SE Reboot. Field introduced in 18.2.5.
	EnablePcapTxRing *bool `json:"enable_pcap_tx_ring,omitempty"`

	// Enable routing for this ServiceEngineGroup . Field deprecated in 18.2.5.
	EnableRouting *bool `json:"enable_routing,omitempty"`

	// Enable VIP on all interfaces of SE. Field deprecated in 18.2.5. Field introduced in 17.1.1.
	EnableVipOnAllInterfaces *bool `json:"enable_vip_on_all_interfaces,omitempty"`

	// Use Virtual MAC address for interfaces on which floating interface IPs are placed. Field deprecated in 18.2.5.
	EnableVMAC *bool `json:"enable_vmac,omitempty"`

	// End local ephemeral port number for outbound connections. Field introduced in 17.2.13, 18.1.5, 18.2.1.
	EphemeralPortrangeEnd *int32 `json:"ephemeral_portrange_end,omitempty"`

	// Start local ephemeral port number for outbound connections. Field introduced in 17.2.13, 18.1.5, 18.2.1.
	EphemeralPortrangeStart *int32 `json:"ephemeral_portrange_start,omitempty"`

	// Multiplier for extra config to support large VS/Pool config.
	ExtraConfigMultiplier *float64 `json:"extra_config_multiplier,omitempty"`

	// Extra config memory to support large Geo DB configuration. Field introduced in 17.1.1. Unit is MB.
	ExtraSharedConfigMemory *int32 `json:"extra_shared_config_memory,omitempty"`

	// If ServiceEngineGroup is configured for Legacy 1+1 Active Standby HA Mode, Floating IP's will be advertised only by the Active SE in the Pair. Virtual Services in this group must be disabled/enabled for any changes to the Floating IP's to take effect. Only active SE hosting VS tagged with Active Standby SE 1 Tag will advertise this floating IP when manual load distribution is enabled. Field deprecated in 18.2.5.
	FloatingIntfIP []*IPAddr `json:"floating_intf_ip,omitempty"`

	// If ServiceEngineGroup is configured for Legacy 1+1 Active Standby HA Mode, Floating IP's will be advertised only by the Active SE in the Pair. Virtual Services in this group must be disabled/enabled for any changes to the Floating IP's to take effect. Only active SE hosting VS tagged with Active Standby SE 2 Tag will advertise this floating IP when manual load distribution is enabled. Field deprecated in 18.2.5.
	FloatingIntfIPSe2 []*IPAddr `json:"floating_intf_ip_se_2,omitempty"`

	// Maximum number of flow table entries that have not completed TCP three-way handshake yet. Field introduced in 17.2.5.
	FlowTableNewSynMaxEntries *int32 `json:"flow_table_new_syn_max_entries,omitempty"`

	// Number of entries in the free list. Field introduced in 17.2.10, 18.1.2.
	FreeListSize *int32 `json:"free_list_size,omitempty"`

	// GratArp periodicity for VIP-IP. Allowed values are 5-30. Field introduced in 18.2.3. Unit is MIN.
	GratarpPermanentPeriodicity *int32 `json:"gratarp_permanent_periodicity,omitempty"`

	// High Availability mode for all the Virtual Services using this Service Engine group. Enum options - HA_MODE_SHARED_PAIR, HA_MODE_SHARED, HA_MODE_LEGACY_ACTIVE_STANDBY.
	HaMode *string `json:"ha_mode,omitempty"`

	//  It is a reference to an object of type HardwareSecurityModuleGroup.
	HardwaresecuritymodulegroupRef *string `json:"hardwaresecuritymodulegroup_ref,omitempty"`

	// Minimum required heap memory to apply any configuration. Allowed values are 0-100. Field introduced in 18.1.2. Unit is MB.
	HeapMinimumConfigMemory *int32 `json:"heap_minimum_config_memory,omitempty"`

	// Enable active health monitoring from the standby SE for all placed virtual services.
	HmOnStandby *bool `json:"hm_on_standby,omitempty"`

	// Key of a (Key, Value) pair identifying a label for a set of Nodes usually in Container Clouds. Needs to be specified together with host_attribute_value. SEs can be configured differently including HA modes across different SE Groups. May also be used for isolation between different classes of VirtualServices. VirtualServices' SE Group may be specified via annotations/labels. A OpenShift/Kubernetes namespace maybe annotated with a matching SE Group label as openshift.io/node-selector  apptype=prod. When multiple SE Groups are used in a Cloud with host attributes specified,just a single SE Group can exist as a match-all SE Group without a host_attribute_key.
	HostAttributeKey *string `json:"host_attribute_key,omitempty"`

	// Value of a (Key, Value) pair identifying a label for a set of Nodes usually in Container Clouds. Needs to be specified together with host_attribute_key.
	HostAttributeValue *string `json:"host_attribute_value,omitempty"`

	// Enable the host gateway monitor when service engine is deployed as docker container. Disabled by default. Field introduced in 17.2.4.
	HostGatewayMonitor *bool `json:"host_gateway_monitor,omitempty"`

	// Override default hypervisor. Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN.
	Hypervisor *string `json:"hypervisor,omitempty"`

	// Ignore RTT samples if it is above threshold. Field introduced in 17.1.6,17.2.2. Unit is MILLISECONDS.
	IgnoreRttThreshold *int32 `json:"ignore_rtt_threshold,omitempty"`

	// Program SE security group ingress rules to allow VIP data access from remote CIDR type. Enum options - SG_INGRESS_ACCESS_NONE, SG_INGRESS_ACCESS_ALL, SG_INGRESS_ACCESS_VPC. Field introduced in 17.1.5.
	IngressAccessData *string `json:"ingress_access_data,omitempty"`

	// Program SE security group ingress rules to allow SSH/ICMP management access from remote CIDR type. Enum options - SG_INGRESS_ACCESS_NONE, SG_INGRESS_ACCESS_ALL, SG_INGRESS_ACCESS_VPC. Field introduced in 17.1.5.
	IngressAccessMgmt *string `json:"ingress_access_mgmt,omitempty"`

	// Instance/Flavor name for SE instance.
	InstanceFlavor *string `json:"instance_flavor,omitempty"`

	// Additional information associated with instance_flavor. Field introduced in 20.1.1.
	InstanceFlavorInfo *CloudFlavor `json:"instance_flavor_info,omitempty"`

	// Iptable Rules.
	Iptables []*IptableRuleSet `json:"iptables,omitempty"`

	// Labels associated with this SE group. Field introduced in 20.1.1.
	Labels []*KeyValue `json:"labels,omitempty"`

	// Select core with least load for new flow.
	LeastLoadCoreSelection *bool `json:"least_load_core_selection,omitempty"`

	// Specifies the license tier which would be used. This field by default inherits the value from cloud. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC. Field introduced in 17.2.5.
	LicenseTier *string `json:"license_tier,omitempty"`

	// If no license type is specified then default license enforcement for the cloud type is chosen. Enum options - LIC_BACKEND_SERVERS, LIC_SOCKETS, LIC_CORES, LIC_HOSTS, LIC_SE_BANDWIDTH, LIC_METERED_SE_BANDWIDTH. Field introduced in 17.2.5.
	LicenseType *string `json:"license_type,omitempty"`

	// Maximum disk capacity (in MB) to be allocated to an SE. This is exclusively used for debug and log data. Unit is MB.
	LogDisksz *int32 `json:"log_disksz,omitempty"`

	// Maximum number of external health monitors that can run concurrently in a service engine. This helps control the CPU and memory use by external health monitors. Special values are 0- 'Value will be internally calculated based on cpu and memory'. Field introduced in 18.2.7.
	MaxConcurrentExternalHm *int32 `json:"max_concurrent_external_hm,omitempty"`

	// When CPU usage on an SE exceeds this threshold, Virtual Services hosted on this SE may be rebalanced to other SEs to reduce load. A new SE may be created as part of this process. Allowed values are 40-90. Unit is PERCENT.
	MaxCPUUsage *int32 `json:"max_cpu_usage,omitempty"`

	// Max bytes that can be allocated in a single mempool. Field introduced in 18.1.5. Unit is MB.
	MaxMemoryPerMempool *int32 `json:"max_memory_per_mempool,omitempty"`

	// Configures the maximum number of se_dp processes created on the SE, requires SE reboot. If not configured, defaults to the number of CPUs on the SE. This should only be used if user wants to limit the number of se_dps to less than the available CPUs on the SE. Allowed values are 1-128. Field introduced in 20.1.1.
	MaxNumSeDps *int32 `json:"max_num_se_dps,omitempty"`

	// Applicable to Azure platform only. Maximum number of public IPs per Azure LB. . Field introduced in 17.2.12, 18.1.2.
	MaxPublicIpsPerLb *int32 `json:"max_public_ips_per_lb,omitempty"`

	// Maximum number of queues per vnic Setting to '0' utilises all queues that are distributed across dispatcher cores. Allowed values are 0,1,2,4,8,16. Field introduced in 18.2.7, 20.1.1.
	MaxQueuesPerVnic *int32 `json:"max_queues_per_vnic,omitempty"`

	// Applicable to Azure platform only. Maximum number of rules per Azure LB. . Field introduced in 17.2.12, 18.1.2.
	MaxRulesPerLb *int32 `json:"max_rules_per_lb,omitempty"`

	// Maximum number of active Service Engines for the Virtual Service. Allowed values are 1-64.
	MaxScaleoutPerVs *int32 `json:"max_scaleout_per_vs,omitempty"`

	// Maximum number of Services Engines in this group. Allowed values are 0-1000.
	MaxSe *int32 `json:"max_se,omitempty"`

	// Maximum number of Virtual Services that can be placed on a single Service Engine. East West Virtual Services are excluded from this limit. Allowed values are 1-1000.
	MaxVsPerSe *int32 `json:"max_vs_per_se,omitempty"`

	// Placeholder for description of property mem_reserve of obj type ServiceEngineGroup field type str  type boolean
	MemReserve *bool `json:"mem_reserve,omitempty"`

	// Indicates the percent of memory reserved for config updates. Allowed values are 0-100. Field introduced in 18.1.2. Unit is PERCENT.
	MemoryForConfigUpdate *int32 `json:"memory_for_config_update,omitempty"`

	// Amount of memory for each of the Service Engine virtual machines.
	MemoryPerSe *int32 `json:"memory_per_se,omitempty"`

	// Management network to use for Avi Service Engines. It is a reference to an object of type Network.
	MgmtNetworkRef *string `json:"mgmt_network_ref,omitempty"`

	// Management subnet to use for Avi Service Engines.
	MgmtSubnet *IPAddrPrefix `json:"mgmt_subnet,omitempty"`

	// When CPU usage on an SE falls below the minimum threshold, Virtual Services hosted on the SE may be consolidated onto other underutilized SEs. After consolidation, unused Service Engines may then be eligible for deletion. . Allowed values are 20-60. Unit is PERCENT.
	MinCPUUsage *int32 `json:"min_cpu_usage,omitempty"`

	// Minimum number of active Service Engines for the Virtual Service. Allowed values are 1-64.
	MinScaleoutPerVs *int32 `json:"min_scaleout_per_vs,omitempty"`

	// Minimum number of Services Engines in this group (relevant for SE AutoRebalance only). Allowed values are 0-1000. Field introduced in 17.2.13,18.1.3,18.2.1.
	MinSe *int32 `json:"min_se,omitempty"`

	// Indicates the percent of memory reserved for connections. Allowed values are 0-100. Field introduced in 18.1.2. Unit is PERCENT.
	MinimumConnectionMemory *int32 `json:"minimum_connection_memory,omitempty"`

	// Required available config memory to apply any configuration. Allowed values are 0-90. Field deprecated in 18.1.2. Field introduced in 18.1.1. Unit is PERCENT.
	MinimumRequiredConfigMemory *int32 `json:"minimum_required_config_memory,omitempty"`

	// Number of threads to use for log streaming. Allowed values are 1-100. Field introduced in 17.2.12, 18.1.2.
	NLogStreamingThreads *int32 `json:"n_log_streaming_threads,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Idle timeout in seconds for nat tcp flows in closed state. Allowed values are 1-3600. Field deprecated in 18.2.5. Field introduced in 18.2.5. Unit is SECONDS.
	NatFlowTCPClosedTimeout *int32 `json:"nat_flow_tcp_closed_timeout,omitempty"`

	// Idle timeout in seconds for nat tcp flows in established state. Allowed values are 1-3600. Field deprecated in 18.2.5. Field introduced in 18.2.5. Unit is SECONDS.
	NatFlowTCPEstablishedTimeout *int32 `json:"nat_flow_tcp_established_timeout,omitempty"`

	// Idle timeout in seconds for nat tcp flows in half closed state. Allowed values are 1-3600. Field deprecated in 18.2.5. Field introduced in 18.2.5. Unit is SECONDS.
	NatFlowTCPHalfClosedTimeout *int32 `json:"nat_flow_tcp_half_closed_timeout,omitempty"`

	// Idle timeout in seconds for nat tcp flows in handshake state. Allowed values are 1-3600. Field deprecated in 18.2.5. Field introduced in 18.2.5. Unit is SECONDS.
	NatFlowTCPHandshakeTimeout *int32 `json:"nat_flow_tcp_handshake_timeout,omitempty"`

	// Idle timeout in seconds for nat udp flows in noresponse state. Allowed values are 1-3600. Field deprecated in 18.2.5. Field introduced in 18.2.5. Unit is SECONDS.
	NatFlowUDPNoresponseTimeout *int32 `json:"nat_flow_udp_noresponse_timeout,omitempty"`

	// Idle timeout in seconds for nat udp flows in response state. Allowed values are 1-3600. Field deprecated in 18.2.5. Field introduced in 18.2.5. Unit is SECONDS.
	NatFlowUDPResponseTimeout *int32 `json:"nat_flow_udp_response_timeout,omitempty"`

	// This setting limits the number of non-significant logs generated per second per core on this SE. Default is 100 logs per second. Set it to zero (0) to disable throttling. Field introduced in 17.1.3. Unit is PER_SECOND.
	NonSignificantLogThrottle *int32 `json:"non_significant_log_throttle,omitempty"`

	// Number of dispatcher cores (0,1,2,4,8 or 16). If set to 0, then number of dispatcher cores is deduced automatically.Requires SE Reboot. Allowed values are 0,1,2,4,8,16. Field introduced in 17.2.12, 18.1.3, 18.2.1.
	NumDispatcherCores *int32 `json:"num_dispatcher_cores,omitempty"`

	// Number of changes in num flow cores sum to ignore.
	NumFlowCoresSumChangesToIgnore *int32 `json:"num_flow_cores_sum_changes_to_ignore,omitempty"`

	//  Field deprecated in 17.1.1.
	OpenstackAvailabilityZone *string `json:"openstack_availability_zone,omitempty"`

	//  Field introduced in 17.1.1.
	OpenstackAvailabilityZones []string `json:"openstack_availability_zones,omitempty"`

	// Avi Management network name.
	OpenstackMgmtNetworkName *string `json:"openstack_mgmt_network_name,omitempty"`

	// Management network UUID.
	OpenstackMgmtNetworkUUID *string `json:"openstack_mgmt_network_uuid,omitempty"`

	// Amount of extra memory to be reserved for use by the Operating System on a Service Engine. Unit is MB.
	OsReservedMemory *int32 `json:"os_reserved_memory,omitempty"`

	// Determines the PCAP transmit mode of operation. Requires SE Reboot. Enum options - PCAP_TX_AUTO, PCAP_TX_SOCKET, PCAP_TX_RING. Field introduced in 18.2.8, 20.1.1.
	PcapTxMode *string `json:"pcap_tx_mode,omitempty"`

	// Per-app SE mode is designed for deploying dedicated load balancers per app (VS). In this mode, each SE is limited to a max of 2 VSs. vCPUs in per-app SEs count towards licensing usage at 25% rate.
	PerApp *bool `json:"per_app,omitempty"`

	// If placement mode is 'Auto', Virtual Services are automatically placed on Service Engines. Enum options - PLACEMENT_MODE_AUTO.
	PlacementMode *string `json:"placement_mode,omitempty"`

	// Enable or disable real time SE metrics.
	RealtimeSeMetrics *MetricsRealTimeUpdate `json:"realtime_se_metrics,omitempty"`

	// Reboot the VM or host on kernel panic. Field introduced in 18.2.5.
	RebootOnPanic *bool `json:"reboot_on_panic,omitempty"`

	// Reboot the system if the SE is stopped. Field deprecated in 18.2.5.
	RebootOnStop *bool `json:"reboot_on_stop,omitempty"`

	// Time interval to re-sync SE's time with wall clock time. Allowed values are 8-600000. Field introduced in 20.1.1. Unit is MILLISECONDS.
	ResyncTimeInterval *int32 `json:"resync_time_interval,omitempty"`

	// Select the SE bandwidth for the bandwidth license. Enum options - SE_BANDWIDTH_UNLIMITED, SE_BANDWIDTH_25M, SE_BANDWIDTH_200M, SE_BANDWIDTH_1000M, SE_BANDWIDTH_10000M. Field introduced in 17.2.5.
	SeBandwidthType *string `json:"se_bandwidth_type,omitempty"`

	// Duration to preserve unused Service Engine virtual machines before deleting them. If traffic to a Virtual Service were to spike up abruptly, this SE would still be available to be utilized again rather than creating a new SE. If this value is set to 0, Controller will never delete any SEs and administrator has to manually cleanup unused SEs. Allowed values are 0-525600. Unit is MIN.
	SeDeprovisionDelay *int32 `json:"se_deprovision_delay,omitempty"`

	// Placeholder for description of property se_dos_profile of obj type ServiceEngineGroup field type str  type object
	SeDosProfile *DosThresholdProfile `json:"se_dos_profile,omitempty"`

	// The highest supported SE-SE Heartbeat protocol version. This version is reported by Secondary SE to Primary SE in Heartbeat response messages. Allowed values are 1-2. Field introduced in 20.1.1.
	SeDpMaxHbVersion *int32 `json:"se_dp_max_hb_version,omitempty"`

	// Time (in seconds) service engine waits for after generating a Vnic transmit queue stall event before resetting theNIC. Field introduced in 18.2.5.
	SeDpVnicQueueStallEventSleep *int32 `json:"se_dp_vnic_queue_stall_event_sleep,omitempty"`

	// Number of consecutive transmit failures to look for before generating a Vnic transmit queue stall event. Field introduced in 18.2.5.
	SeDpVnicQueueStallThreshold *int32 `json:"se_dp_vnic_queue_stall_threshold,omitempty"`

	// Time (in milliseconds) to wait for network/NIC recovery on detecting a transmit queue stall after which service engine resets the NIC. Field introduced in 18.2.5.
	SeDpVnicQueueStallTimeout *int32 `json:"se_dp_vnic_queue_stall_timeout,omitempty"`

	// Number of consecutive transmit queue stall events in se_dp_vnic_stall_se_restart_window to look for before restarting SE. Field introduced in 18.2.5.
	SeDpVnicRestartOnQueueStallCount *int32 `json:"se_dp_vnic_restart_on_queue_stall_count,omitempty"`

	// Window of time (in seconds) during which se_dp_vnic_restart_on_queue_stall_count number of consecutive stalls results in a SE restart. Field introduced in 18.2.5.
	SeDpVnicStallSeRestartWindow *int32 `json:"se_dp_vnic_stall_se_restart_window,omitempty"`

	// Determines if DPDK pool mode driver should be used or not   0  Automatically determine based on hypervisor/NIC type 1  Unconditionally use DPDK poll mode driver 2  Don't use DPDK poll mode driver.Requires SE Reboot. Allowed values are 0-2. Field introduced in 18.1.3.
	SeDpdkPmd *int32 `json:"se_dpdk_pmd,omitempty"`

	// Flow probe retry count if no replies are received.Requires SE Reboot. Allowed values are 0-5. Field introduced in 18.1.4, 18.2.1.
	SeFlowProbeRetries *int32 `json:"se_flow_probe_retries,omitempty"`

	// Timeout in milliseconds for flow probe retries.Requires SE Reboot. Allowed values are 20-50. Field introduced in 18.2.5. Unit is MILLISECONDS.
	SeFlowProbeRetryTimer *int32 `json:"se_flow_probe_retry_timer,omitempty"`

	// Timeout in milliseconds for flow probe entries. Allowed values are 10-200. Field deprecated in 18.2.5. Field introduced in 18.1.4, 18.2.1. Unit is MILLISECONDS.
	SeFlowProbeTimer *int32 `json:"se_flow_probe_timer,omitempty"`

	// Controls the distribution of SE data path processes on CPUs which support hyper-threading. Requires hyper-threading to be enabled at host level. Requires SE Reboot. For more details please refer to SE placement KB. Enum options - SE_CPU_HT_AUTO, SE_CPU_HT_SPARSE_DISPATCHER_PRIORITY, SE_CPU_HT_SPARSE_PROXY_PRIORITY, SE_CPU_HT_PACKED_CORES. Field introduced in 20.1.1.
	SeHyperthreadedMode *string `json:"se_hyperthreaded_mode,omitempty"`

	// UDP Port for SE_DP IPC in Docker bridge mode. Field deprecated in 20.1.1. Field introduced in 17.1.2.
	SeIpcUDPPort *int32 `json:"se_ipc_udp_port,omitempty"`

	// Knob to control burst size used in polling KNI interfaces for traffic sent from KNI towards DPDK application Also controls burst size used by KNI module to read pkts punted from DPDK application towards KNI Helps minimize drops in non-VIP traffic in either pathFactor of (0-2) multiplies/divides burst size by 2^N. Allowed values are 0-2. Field introduced in 18.2.6.
	SeKniBurstFactor *int32 `json:"se_kni_burst_factor,omitempty"`

	// Enable or disable Large Receive Optimization for vnics. Requires SE Reboot. Field introduced in 18.2.5.
	SeLro *bool `json:"se_lro,omitempty"`

	// MTU for the VNICs of SEs in the SE group. Allowed values are 512-9000. Field introduced in 18.2.8, 20.1.1.
	SeMtu *int32 `json:"se_mtu,omitempty"`

	// Prefix to use for virtual machine name of Service Engines.
	SeNamePrefix *string `json:"se_name_prefix,omitempty"`

	// Enables lookahead mode of packet receive in PCAP mode. Introduced to overcome an issue with hv_netvsc driver. Lookahead mode attempts to ensure that application and kernel's view of the receive rings are consistent. Field introduced in 18.2.3.
	SePcapLookahead *bool `json:"se_pcap_lookahead,omitempty"`

	// Max number of packets the pcap interface can hold and if the value is 0 the optimum value will be chosen. The optimum value will be chosen based on SE-memory, Cloud Type and Number of Interfaces.Requires SE Reboot. Field introduced in 18.2.5.
	SePcapPktCount *int32 `json:"se_pcap_pkt_count,omitempty"`

	// Max size of each packet in the pcap interface. Requires SE Reboot. Field introduced in 18.2.5.
	SePcapPktSz *int32 `json:"se_pcap_pkt_sz,omitempty"`

	// Bypass the kernel's traffic control layer, to deliver packets directly to the driver. Enabling this feature results in egress packets not being captured in host tcpdump. Note   brief packet reordering or loss may occur upon toggle. Field introduced in 18.2.6.
	SePcapQdiscBypass *bool `json:"se_pcap_qdisc_bypass,omitempty"`

	// Frequency in seconds at which periodically a PCAP reinit check is triggered. May be used in conjunction with the configuration pcap_reinit_threshold. (Valid range   15 mins - 12 hours, 0 - disables). Allowed values are 900-43200. Special values are 0- 'disable'. Field introduced in 17.2.13, 18.1.3, 18.2.1. Unit is SEC.
	SePcapReinitFrequency *int32 `json:"se_pcap_reinit_frequency,omitempty"`

	// Threshold for input packet receive errors in PCAP mode exceeding which a PCAP reinit is triggered. If not set, an unconditional reinit is performed. This value is checked every pcap_reinit_frequency interval. Field introduced in 17.2.13, 18.1.3, 18.2.1. Unit is METRIC_COUNT.
	SePcapReinitThreshold *int32 `json:"se_pcap_reinit_threshold,omitempty"`

	// TCP port on SE where echo service will be run. Field introduced in 17.2.2.
	SeProbePort *int32 `json:"se_probe_port,omitempty"`

	// UDP Port for punted packets in Docker bridge mode. Field deprecated in 20.1.1. Field introduced in 17.1.2.
	SeRemotePuntUDPPort *int32 `json:"se_remote_punt_udp_port,omitempty"`

	// Rate limiter properties. Field introduced in 20.1.1.
	SeRlProp *RateLimiterProperties `json:"se_rl_prop,omitempty"`

	// Enable routing via Service Engine Datapath. When disabled, routing is done by the Linux kernel. IP Routing needs to be enabled in Service Engine Group for SE Routing to be effective. Field deprecated in 18.2.5. Field introduced in 18.2.3.
	SeRouting *bool `json:"se_routing,omitempty"`

	// Minimum time to wait on server between taking sampleswhen sampling the navigation timing data from the end user client. Field introduced in 18.2.6. Unit is SEC.
	SeRumSamplingNavInterval *int32 `json:"se_rum_sampling_nav_interval,omitempty"`

	// Percentage of navigation timing data from the end user client, used for sampling to get client insights. Field introduced in 18.2.6.
	SeRumSamplingNavPercent *int32 `json:"se_rum_sampling_nav_percent,omitempty"`

	// Minimum time to wait on server between taking sampleswhen sampling the resource timing data from the end user client. Field introduced in 18.2.6. Unit is SEC.
	SeRumSamplingResInterval *int32 `json:"se_rum_sampling_res_interval,omitempty"`

	// Percentage of resource timing data from the end user client used for sampling to get client insight. Field introduced in 18.2.6.
	SeRumSamplingResPercent *int32 `json:"se_rum_sampling_res_percent,omitempty"`

	// Sideband traffic will be handled by a dedicated core.Requires SE Reboot. Field introduced in 16.5.2, 17.1.9, 17.2.3.
	SeSbDedicatedCore *bool `json:"se_sb_dedicated_core,omitempty"`

	// Number of Sideband threads per SE.Requires SE Reboot. Allowed values are 1-128. Field introduced in 16.5.2, 17.1.9, 17.2.3.
	SeSbThreads *int32 `json:"se_sb_threads,omitempty"`

	// Multiplier for SE threads based on vCPU. Allowed values are 1-10.
	SeThreadMultiplier *int32 `json:"se_thread_multiplier,omitempty"`

	// Traceroute port range. Field introduced in 17.2.8.
	SeTracertPortRange *PortRange `json:"se_tracert_port_range,omitempty"`

	// Determines if DSR from secondary SE is active or not  0  Automatically determine based on hypervisor type. 1  Disable DSR unconditionally. 2  Enable DSR unconditionally. Allowed values are 0-2. Field introduced in 17.1.1.
	SeTunnelMode *int32 `json:"se_tunnel_mode,omitempty"`

	// UDP Port for tunneled packets from secondary to primary SE in Docker bridge mode.Requires SE Reboot. Field introduced in 17.1.3.
	SeTunnelUDPPort *int32 `json:"se_tunnel_udp_port,omitempty"`

	// Number of packets to batch for transmit to the nic. Requires SE Reboot. Field introduced in 18.2.5.
	SeTxBatchSize *int32 `json:"se_tx_batch_size,omitempty"`

	// Determines if SE-SE IPC messages are encapsulated in a UDP header  0  Automatically determine based on hypervisor type. 1  Use UDP encap unconditionally.Requires SE Reboot. Allowed values are 0-1. Field introduced in 17.1.2.
	SeUDPEncapIpc *int32 `json:"se_udp_encap_ipc,omitempty"`

	// Determines if DPDK library should be used or not   0  Automatically determine based on hypervisor type 1  Use DPDK if PCAP is not enabled 2  Don't use DPDK. Allowed values are 0-2. Field introduced in 18.1.3.
	SeUseDpdk *int32 `json:"se_use_dpdk,omitempty"`

	// Configure the frequency in milliseconds of software transmit spillover queue flush when enabled. This is necessary to flush any packets in the spillover queue in the absence of a packet transmit in the normal course of operation. Allowed values are 50-500. Special values are 0- 'disable'. Field introduced in 20.1.1. Unit is MILLISECONDS.
	SeVnicTxSwQueueFlushFrequency *int32 `json:"se_vnic_tx_sw_queue_flush_frequency,omitempty"`

	// Configure the size of software transmit spillover queue when enabled. Requires SE Reboot. Allowed values are 128-2048. Field introduced in 20.1.1.
	SeVnicTxSwQueueSize *int32 `json:"se_vnic_tx_sw_queue_size,omitempty"`

	// Maximum number of aggregated vs heartbeat packets to send in a batch. Allowed values are 1-256. Field introduced in 17.1.1.
	SeVsHbMaxPktsInBatch *int32 `json:"se_vs_hb_max_pkts_in_batch,omitempty"`

	// Maximum number of virtualservices for which heartbeat messages are aggregated in one packet. Allowed values are 1-1024. Field introduced in 17.1.1.
	SeVsHbMaxVsInPkt *int32 `json:"se_vs_hb_max_vs_in_pkt,omitempty"`

	// Enable SEs to elect a primary amongst themselves in the absence of a connectivity to controller. Field introduced in 18.1.2.
	SelfSeElection *bool `json:"self_se_election,omitempty"`

	// IPv6 Subnets assigned to the SE group. Required for VS group placement. Field introduced in 18.1.1.
	ServiceIp6Subnets []*IPAddrPrefix `json:"service_ip6_subnets,omitempty"`

	// Subnets assigned to the SE group. Required for VS group placement. Field introduced in 17.1.1.
	ServiceIPSubnets []*IPAddrPrefix `json:"service_ip_subnets,omitempty"`

	// Minimum required shared memory to apply any configuration. Allowed values are 0-100. Field introduced in 18.1.2. Unit is MB.
	ShmMinimumConfigMemory *int32 `json:"shm_minimum_config_memory,omitempty"`

	// This setting limits the number of significant logs generated per second per core on this SE. Default is 100 logs per second. Set it to zero (0) to disable throttling. Field introduced in 17.1.3. Unit is PER_SECOND.
	SignificantLogThrottle *int32 `json:"significant_log_throttle,omitempty"`

	// (Beta) Preprocess SSL Client Hello for SNI hostname extension.If set to True, this will apply SNI child's SSL protocol(s), if they are different from SNI Parent's allowed SSL protocol(s). Field introduced in 17.2.12, 18.1.3.
	SslPreprocessSniHostname *bool `json:"ssl_preprocess_sni_hostname,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// The threshold for the transient shared config memory in the SE. Allowed values are 0-100. Field introduced in 20.1.1. Unit is PERCENT.
	TransientSharedMemoryMax *int32 `json:"transient_shared_memory_max,omitempty"`

	// This setting limits the number of UDF logs generated per second per core on this SE. UDF logs are generated due to the configured client log filters or the rules with logging enabled. Default is 100 logs per second. Set it to zero (0) to disable throttling. Field introduced in 17.1.3. Unit is PER_SECOND.
	UdfLogThrottle *int32 `json:"udf_log_throttle,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Enables the use of hyper-threaded cores on SE. Requires SE Reboot. Field introduced in 20.1.1.
	UseHyperthreadedCores *bool `json:"use_hyperthreaded_cores,omitempty"`

	// Use Standard SKU Azure Load Balancer. By default cloud level flag is set. If not set, it inherits/uses the use_standard_alb flag from the cloud. Field introduced in 18.2.3.
	UseStandardAlb *bool `json:"use_standard_alb,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Placeholder for description of property vcenter_clusters of obj type ServiceEngineGroup field type str  type object
	VcenterClusters *VcenterClusters `json:"vcenter_clusters,omitempty"`

	//  Enum options - VCENTER_DATASTORE_ANY, VCENTER_DATASTORE_LOCAL, VCENTER_DATASTORE_SHARED.
	VcenterDatastoreMode *string `json:"vcenter_datastore_mode,omitempty"`

	// Placeholder for description of property vcenter_datastores of obj type ServiceEngineGroup field type str  type object
	VcenterDatastores []*VcenterDatastore `json:"vcenter_datastores,omitempty"`

	// Placeholder for description of property vcenter_datastores_include of obj type ServiceEngineGroup field type str  type boolean
	VcenterDatastoresInclude *bool `json:"vcenter_datastores_include,omitempty"`

	// Folder to place all the Service Engine virtual machines in vCenter.
	VcenterFolder *string `json:"vcenter_folder,omitempty"`

	// Placeholder for description of property vcenter_hosts of obj type ServiceEngineGroup field type str  type object
	VcenterHosts *VcenterHosts `json:"vcenter_hosts,omitempty"`

	// VCenter information for scoping at Host/Cluster level. Field introduced in 20.1.1.
	Vcenters []*PlacementScopeConfig `json:"vcenters,omitempty"`

	// Number of vcpus for each of the Service Engine virtual machines.
	VcpusPerSe *int32 `json:"vcpus_per_se,omitempty"`

	// When vip_asg is set, Vip configuration will be managed by Avi.User will be able to configure vip_asg or Vips individually at the time of create. Field introduced in 17.2.12, 18.1.2.
	VipAsg *VipAutoscaleGroup `json:"vip_asg,omitempty"`

	// Ensure primary and secondary Service Engines are deployed on different physical hosts.
	VsHostRedundancy *bool `json:"vs_host_redundancy,omitempty"`

	// Time to wait for the scaled in SE to drain existing flows before marking the scalein done. Unit is SEC.
	VsScaleinTimeout *int32 `json:"vs_scalein_timeout,omitempty"`

	// During SE upgrade, Time to wait for the scaled-in SE to drain existing flows before marking the scalein done. Unit is SEC.
	VsScaleinTimeoutForUpgrade *int32 `json:"vs_scalein_timeout_for_upgrade,omitempty"`

	// Time to wait for the scaled out SE to become ready before marking the scaleout done. Unit is SEC.
	VsScaleoutTimeout *int32 `json:"vs_scaleout_timeout,omitempty"`

	// Wait time for sending scaleout ready notification after Virtual Service is marked UP. In certain deployments, there may be an additional delay to accept traffic. For example, for BGP, some time is needed for route advertisement. Allowed values are 0-20. Field introduced in 18.1.5,18.2.1. Unit is SEC.
	VsSeScaleoutAdditionalWaitTime *int32 `json:"vs_se_scaleout_additional_wait_time,omitempty"`

	// Timeout in seconds for Service Engine to sendScaleout Ready notification of a Virtual Service. Allowed values are 0-90. Field introduced in 18.1.5,18.2.1. Unit is SEC.
	VsSeScaleoutReadyTimeout *int32 `json:"vs_se_scaleout_ready_timeout,omitempty"`

	// During SE upgrade in a legacy active/standby segroup, Time to wait for the new primary SE to accept flows before marking the switchover done. Field introduced in 17.2.13,18.1.4,18.2.1. Unit is SEC.
	VsSwitchoverTimeout *int32 `json:"vs_switchover_timeout,omitempty"`

	// Parameters to place Virtual Services on only a subset of the cores of an SE. Field introduced in 17.2.5.
	VssPlacement *VssPlacement `json:"vss_placement,omitempty"`

	// If set, Virtual Services will be placed on only a subset of the cores of an SE. Field introduced in 18.1.1.
	VssPlacementEnabled *bool `json:"vss_placement_enabled,omitempty"`

	// Frequency with which SE publishes WAF learning. Allowed values are 1-43200. Field deprecated in 18.2.3. Field introduced in 18.1.2. Unit is MIN.
	WafLearningInterval *int32 `json:"waf_learning_interval,omitempty"`

	// Amount of memory reserved on SE for WAF learning. This can be atmost 5% of SE memory. Field deprecated in 18.2.3. Field introduced in 18.1.2. Unit is MB.
	WafLearningMemory *int32 `json:"waf_learning_memory,omitempty"`

	// Enable memory pool for WAF.Requires SE Reboot. Field introduced in 17.2.3.
	WafMempool *bool `json:"waf_mempool,omitempty"`

	// Memory pool size used for WAF.Requires SE Reboot. Field introduced in 17.2.3. Unit is KB.
	WafMempoolSize *int32 `json:"waf_mempool_size,omitempty"`
}
