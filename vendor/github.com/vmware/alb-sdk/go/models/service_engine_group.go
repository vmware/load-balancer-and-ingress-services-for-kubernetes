// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServiceEngineGroup service engine group
// swagger:model ServiceEngineGroup
type ServiceEngineGroup struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Enable accelerated networking option for Azure SE. Accelerated networking enables single root I/O virtualization (SR-IOV) to a SE VM. This improves networking performance. Field introduced in 17.2.14,18.1.5,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AcceleratedNetworking *bool `json:"accelerated_networking,omitempty"`

	// Service Engines in active/standby mode for HA failover. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ActiveStandby *bool `json:"active_standby,omitempty"`

	// Enable aggressive failover configuration for ha. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AggressiveFailureDetection *bool `json:"aggressive_failure_detection,omitempty"`

	// In compact placement, Virtual Services are placed on existing SEs until max_vs_per_se limit is reached. In distributed placement, Virtual Services are placed on new SEs until max_se limit is reached. Once this limit is reached, Virtual Services are placed on SEs with least load. Enum options - PLACEMENT_ALGO_PACKED, PLACEMENT_ALGO_DISTRIBUTED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Algo *string `json:"algo,omitempty"`

	// Allow SEs to be created using burst license. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowBurst *bool `json:"allow_burst,omitempty"`

	// A percent value of total SE memory reserved for applicationcaching. This is an SE bootup property and requires SE restart.Requires SE Reboot. Allowed values are 0 - 100. Special values are 0- disable. Field introduced in 18.2.3. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition. Special default for Essentials edition is 0, Basic edition is 0, Enterprise is 10.
	AppCachePercent *uint32 `json:"app_cache_percent,omitempty"`

	// The max memory that can be allocated for the app cache. This value will act as an upper bound on the cache size specified in app_cache_percent. Special values are 0- disable. Field introduced in 20.1.1. Unit is GB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppCacheThreshold *uint32 `json:"app_cache_threshold,omitempty"`

	// A percent value of total SE memory reserved for Application learning. This is an SE bootup property and requires SE restart. Allowed values are 0 - 10. Field introduced in 18.2.3. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppLearningMemoryPercent *uint32 `json:"app_learning_memory_percent,omitempty"`

	// Amount of SE memory in GB until which shared memory is collected in core archive. Field introduced in 17.1.3. Unit is GB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ArchiveShmLimit *uint32 `json:"archive_shm_limit,omitempty"`

	// SSL handshakes will be handled by dedicated SSL Threads.Requires SE Reboot. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AsyncSsl *bool `json:"async_ssl,omitempty"`

	// Number of Async SSL threads per se_dp.Requires SE Reboot. Allowed values are 1-16. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AsyncSslThreads *uint32 `json:"async_ssl_threads,omitempty"`

	// If set, Virtual Services will be automatically migrated when load on an SE is less than minimum or more than maximum thresholds. Only Alerts are generated when the auto_rebalance is not set. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AutoRebalance *bool `json:"auto_rebalance,omitempty"`

	// Capacities of SE for auto rebalance for each criteria. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AutoRebalanceCapacityPerSe []int64 `json:"auto_rebalance_capacity_per_se,omitempty,omitempty"`

	// Set of criteria for SE Auto Rebalance. Enum options - SE_AUTO_REBALANCE_CPU, SE_AUTO_REBALANCE_PPS, SE_AUTO_REBALANCE_MBPS, SE_AUTO_REBALANCE_OPEN_CONNS, SE_AUTO_REBALANCE_CPS. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AutoRebalanceCriteria []string `json:"auto_rebalance_criteria,omitempty"`

	// Frequency of rebalance, if 'Auto rebalance' is enabled. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AutoRebalanceInterval *int32 `json:"auto_rebalance_interval,omitempty"`

	// Redistribution of virtual services from the takeover SE to the replacement SE can cause momentary traffic loss. If the auto-redistribute load option is left in its default off state, any desired rebalancing requires calls to REST API. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AutoRedistributeActiveStandbyLoad *bool `json:"auto_redistribute_active_standby_load,omitempty"`

	// Availability zones for Virtual Service High Availability. It is a reference to an object of type AvailabilityZone. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvailabilityZoneRefs []string `json:"availability_zone_refs,omitempty"`

	// Control if dispatcher core also handles TCP flows in baremetal SE. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	BaremetalDispatcherHandlesFlows *bool `json:"baremetal_dispatcher_handles_flows,omitempty"`

	// Enable BGP peer monitoring based failover. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BgpPeerMonitorFailoverEnabled *bool `json:"bgp_peer_monitor_failover_enabled,omitempty"`

	// BGP peer state update interval. Allowed values are 5-100. Field introduced in 17.2.14,18.1.5,18.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BgpStateUpdateInterval *uint32 `json:"bgp_state_update_interval,omitempty"`

	// Excess Service Engine capacity provisioned for HA failover. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BufferSe *int32 `json:"buffer_se,omitempty"`

	//  It is a reference to an object of type Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudRef *string `json:"cloud_ref,omitempty"`

	// Compress IP rules into a single subnet based IP rule for each north-south IPAM subnet configured in PCAP mode in OpenShift/Kubernetes node. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CompressIPRulesForEachNsSubnet *bool `json:"compress_ip_rules_for_each_ns_subnet,omitempty"`

	// Enable config debugs on all cores of SE. Field introduced in 17.2.13,18.1.5,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigDebugsOnAllCores *bool `json:"config_debugs_on_all_cores,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Percentage of memory for connection state. This will come at the expense of memory used for HTTP in-memory cache. Allowed values are 10-90. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConnectionMemoryPercentage *uint32 `json:"connection_memory_percentage,omitempty"`

	// Include shared memory for app cache in core file.Requires SE Reboot. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CoreShmAppCache *bool `json:"core_shm_app_cache,omitempty"`

	// Include shared memory for app learning in core file.Requires SE Reboot. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CoreShmAppLearning *bool `json:"core_shm_app_learning,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CPUReserve *bool `json:"cpu_reserve,omitempty"`

	// Allocate all the CPU cores for the Service Engine Virtual Machines  on the same CPU socket. Applicable only for vCenter Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CPUSocketAffinity *bool `json:"cpu_socket_affinity,omitempty"`

	// Custom Security Groups to be associated with data vNics for SE instances in OpenStack and AWS Clouds. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CustomSecuritygroupsData []string `json:"custom_securitygroups_data,omitempty"`

	// Custom Security Groups to be associated with management vNic for SE instances in OpenStack and AWS Clouds. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CustomSecuritygroupsMgmt []string `json:"custom_securitygroups_mgmt,omitempty"`

	// Custom tag will be used to create the tags for SE instance in AWS. Note this is not the same as the prefix for SE name. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CustomTag []*CustomTag `json:"custom_tag,omitempty"`

	// Subnet used to spin up the data nic for Service Engines, used only for Azure cloud. Overrides the cloud level setting for Service Engine subnet. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DataNetworkID *string `json:"data_network_id,omitempty"`

	// Number of instructions before datascript times out. Allowed values are 0-100000000. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DatascriptTimeout *uint64 `json:"datascript_timeout,omitempty"`

	// If activated, IPv6 address and route discovery are deactivated.Requires SE reboot. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeactivateIPV6Discovery *bool `json:"deactivate_ipv6_discovery,omitempty"`

	// Deactivate filtering of packets to KNI interface. To be used under surveillance of Avi Support. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeactivateKniFilteringAtDispatcher *bool `json:"deactivate_kni_filtering_at_dispatcher,omitempty"`

	// Dedicate the core that handles packet receive/transmit from the network to just the dispatching function. Don't use it for TCP/IP and SSL functions. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DedicatedDispatcherCore *bool `json:"dedicated_dispatcher_core,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// By default, Avi creates and manages security groups along with custom sg provided by user. Set this to True to disallow Avi to create and manage new security groups. Avi will only make use of custom security groups provided by user. This option is supported for AWS and OpenStack cloud types. Field introduced in 17.2.13,18.1.4,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableAviSecuritygroups *bool `json:"disable_avi_securitygroups,omitempty"`

	// Stop using TCP/UDP and IP checksum offload features of NICs. Field introduced in 17.1.14, 17.2.5, 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableCsumOffloads *bool `json:"disable_csum_offloads,omitempty"`

	// Disable Flow Probes for Scaled out VS'es. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DisableFlowProbes *bool `json:"disable_flow_probes,omitempty"`

	// Disable Generic Receive Offload (GRO) in DPDK poll-mode driver packet receive path.  GRO can be enabled on NICs that do not support LRO (Large Receive Offload) or do not gain performance boost from LRO. GRO is on by default on NICs in a system with 8 vCPUs or higher. Field introduced in 17.2.5, 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableGro *bool `json:"disable_gro,omitempty"`

	// If set, disable the config memory check done in service engine. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableSeMemoryCheck *bool `json:"disable_se_memory_check,omitempty"`

	// Disable TCP Segmentation Offload (TSO) in DPDK poll-mode driver packet transmit path. TSO is on by default on NICs that support it. Field introduced in 17.2.5, 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableTso *bool `json:"disable_tso,omitempty"`

	// Amount of disk space for each of the Service Engine virtual machines. Unit is GB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DiskPerSe *int32 `json:"disk_per_se,omitempty"`

	// Use both the active and standby Service Engines for Virtual Service placement in the legacy active standby HA mode. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	DistributeLoadActiveStandby *bool `json:"distribute_load_active_standby,omitempty"`

	// Distributes queue ownership among cores so multiple cores handle dispatcher duties. Requires SE Reboot. Deprecated from 18.2.8, instead use max_queues_per_vnic. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	DistributeQueues *bool `json:"distribute_queues,omitempty"`

	// Distributes vnic ownership among cores so multiple cores handle dispatcher duties.Requires SE Reboot. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	DistributeVnics *bool `json:"distribute_vnics,omitempty"`

	// Timeout for downstream to become writable. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DownstreamSendTimeout *uint32 `json:"downstream_send_timeout,omitempty"`

	// Dequeue interval for receive queue from se_dp in aggressive mode. Allowed values are 1-1000. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	DpAggressiveDeqIntervalMsec *uint32 `json:"dp_aggressive_deq_interval_msec,omitempty"`

	// Enqueue interval for request queue to se_dp in aggressive mode. Allowed values are 1-1000. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	DpAggressiveEnqIntervalMsec *uint32 `json:"dp_aggressive_enq_interval_msec,omitempty"`

	// Frequency of SE - SE HB messages when aggressive failure mode detection is enabled. Field introduced in 20.1.3. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DpAggressiveHbFrequency *uint32 `json:"dp_aggressive_hb_frequency,omitempty"`

	// Consecutive HB failures after which failure is reported to controller,when aggressive failure mode detection is enabled. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DpAggressiveHbTimeoutCount *uint32 `json:"dp_aggressive_hb_timeout_count,omitempty"`

	// Dequeue interval for receive queue from se_dp. Allowed values are 1-1000. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	DpDeqIntervalMsec *uint32 `json:"dp_deq_interval_msec,omitempty"`

	// Enqueue interval for request queue to se_dp. Allowed values are 1-1000. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	DpEnqIntervalMsec *uint32 `json:"dp_enq_interval_msec,omitempty"`

	// Frequency of SE - SE HB messages when aggressive failure mode detection is not enabled. Field introduced in 20.1.3. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DpHbFrequency *uint32 `json:"dp_hb_frequency,omitempty"`

	// Consecutive HB failures after which failure is reported to controller, when aggressive failure mode detection is not enabled. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DpHbTimeoutCount *uint32 `json:"dp_hb_timeout_count,omitempty"`

	// The timeout for GRO coalescing interval. 0 indicates non-timer based GRO. Allowed values are 0-900. Field introduced in 22.1.1. Unit is MICROSECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DpdkGroTimeoutInterval *uint32 `json:"dpdk_gro_timeout_interval,omitempty"`

	// Enable GratArp for VIP_IP. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableGratarpPermanent *bool `json:"enable_gratarp_permanent,omitempty"`

	// Enable HSM luna engine logs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableHsmLog *bool `json:"enable_hsm_log,omitempty"`

	// (This is a beta feature). Enable HSM key priming. If enabled, key handles on the hsm will be synced to SE before processing client connections. Field introduced in 17.2.7, 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableHsmPriming *bool `json:"enable_hsm_priming,omitempty"`

	// Applicable only for Azure cloud with Basic SKU LB. If set, additional Azure LBs will be automatically created if resources in existing LB are exhausted. Field introduced in 17.2.10, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableMultiLb *bool `json:"enable_multi_lb,omitempty"`

	// Enable TX ring support in pcap mode of operation. TSO feature is not supported with TX Ring enabled. Deprecated from 18.2.8, instead use pcap_tx_mode. Requires SE Reboot. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnablePcapTxRing *bool `json:"enable_pcap_tx_ring,omitempty"`

	// End local ephemeral port number for outbound connections. Field introduced in 17.2.13, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EphemeralPortrangeEnd *uint32 `json:"ephemeral_portrange_end,omitempty"`

	// Start local ephemeral port number for outbound connections. Field introduced in 17.2.13, 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EphemeralPortrangeStart *uint32 `json:"ephemeral_portrange_start,omitempty"`

	// Multiplier for extra config to support large VS/Pool config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExtraConfigMultiplier *float64 `json:"extra_config_multiplier,omitempty"`

	// Extra config memory to support large Geo DB configuration. Field introduced in 17.1.1. Unit is MB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ExtraSharedConfigMemory *uint32 `json:"extra_shared_config_memory,omitempty"`

	// Maximum number of flow table entries that have not completed TCP three-way handshake yet. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FlowTableNewSynMaxEntries *uint32 `json:"flow_table_new_syn_max_entries,omitempty"`

	// Number of entries in the free list. Field introduced in 17.2.10, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FreeListSize *uint32 `json:"free_list_size,omitempty"`

	// Google Cloud Platform, Service Engine Group Configuration. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GcpConfig *GCPSeGroupConfig `json:"gcp_config,omitempty"`

	// GratArp periodicity for VIP-IP. Allowed values are 5-30. Field introduced in 18.2.3. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GratarpPermanentPeriodicity *uint32 `json:"gratarp_permanent_periodicity,omitempty"`

	// Timeout in seconds that SE waits for a grpc channel to connect to server, before it retries. Allowed values are 5-45. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GrpcChannelConnectTimeout *uint32 `json:"grpc_channel_connect_timeout,omitempty"`

	// High Availability mode for all the Virtual Services using this Service Engine group. Enum options - HA_MODE_SHARED_PAIR, HA_MODE_SHARED, HA_MODE_LEGACY_ACTIVE_STANDBY. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- HA_MODE_LEGACY_ACTIVE_STANDBY), Basic edition(Allowed values- HA_MODE_LEGACY_ACTIVE_STANDBY), Enterprise with Cloud Services edition. Special default for Essentials edition is HA_MODE_LEGACY_ACTIVE_STANDBY, Basic edition is HA_MODE_LEGACY_ACTIVE_STANDBY, Enterprise is HA_MODE_SHARED.
	HaMode *string `json:"ha_mode,omitempty"`

	// Configuration to handle per packet attack handling.For example, DNS Reflection Attack is a type of attack where a response packet is sent to the DNS VS.This configuration tells if such packets should be dropped without further processing. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HandlePerPktAttack *bool `json:"handle_per_pkt_attack,omitempty"`

	//  It is a reference to an object of type HardwareSecurityModuleGroup. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HardwaresecuritymodulegroupRef *string `json:"hardwaresecuritymodulegroup_ref,omitempty"`

	// Minimum required heap memory to apply any configuration. Allowed values are 0-100. Field introduced in 18.1.2. Unit is MB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeapMinimumConfigMemory *uint32 `json:"heap_minimum_config_memory,omitempty"`

	// Enable active health monitoring from the standby SE for all placed virtual services. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition. Special default for Essentials edition is false, Basic edition is false, Enterprise is True.
	HmOnStandby *bool `json:"hm_on_standby,omitempty"`

	// Key of a (Key, Value) pair identifying a label for a set of Nodes usually in Container Clouds. Needs to be specified together with host_attribute_value. SEs can be configured differently including HA modes across different SE Groups. May also be used for isolation between different classes of VirtualServices. VirtualServices' SE Group may be specified via annotations/labels. A OpenShift/Kubernetes namespace maybe annotated with a matching SE Group label as openshift.io/node-selector  apptype=prod. When multiple SE Groups are used in a Cloud with host attributes specified,just a single SE Group can exist as a match-all SE Group without a host_attribute_key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostAttributeKey *string `json:"host_attribute_key,omitempty"`

	// Value of a (Key, Value) pair identifying a label for a set of Nodes usually in Container Clouds. Needs to be specified together with host_attribute_key. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostAttributeValue *string `json:"host_attribute_value,omitempty"`

	// Enable the host gateway monitor when service engine is deployed as docker container. Disabled by default. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostGatewayMonitor *bool `json:"host_gateway_monitor,omitempty"`

	// Enable Javascript console logs on the client browser when collecting client insights. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	HTTPRumConsoleLog *bool `json:"http_rum_console_log,omitempty"`

	// Minimum response size content length to sample for client insights. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 64), Basic edition(Allowed values- 64), Enterprise with Cloud Services edition.
	HTTPRumMinContentLength *uint32 `json:"http_rum_min_content_length,omitempty"`

	// Toggles SE hybrid only mode of operation in DPDK mode with RSS configured;where-in each SE datapath instance operates as a standalone hybrid instance performing both dispatcher and proxy function. Requires reboot. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HybridRssMode *bool `json:"hybrid_rss_mode,omitempty"`

	// Override default hypervisor. Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Hypervisor *string `json:"hypervisor,omitempty"`

	// Ignore docker mac change. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	IgnoreDockerMacChange *bool `json:"ignore_docker_mac_change,omitempty"`

	// Ignore RTT samples if it is above threshold. Field introduced in 17.1.6,17.2.2. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IgnoreRttThreshold *uint32 `json:"ignore_rtt_threshold,omitempty"`

	// Program SE security group ingress rules to allow VIP data access from remote CIDR type. Enum options - SG_INGRESS_ACCESS_NONE, SG_INGRESS_ACCESS_ALL, SG_INGRESS_ACCESS_VPC. Field introduced in 17.1.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IngressAccessData *string `json:"ingress_access_data,omitempty"`

	// Program SE security group ingress rules to allow SSH/ICMP management access from remote CIDR type. Enum options - SG_INGRESS_ACCESS_NONE, SG_INGRESS_ACCESS_ALL, SG_INGRESS_ACCESS_VPC. Field introduced in 17.1.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IngressAccessMgmt *string `json:"ingress_access_mgmt,omitempty"`

	// Instance/Flavor name for SE instance. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InstanceFlavor *string `json:"instance_flavor,omitempty"`

	// Additional information associated with instance_flavor. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	InstanceFlavorInfo *CloudFlavor `json:"instance_flavor_info,omitempty"`

	// Iptable Rules. Maximum of 128 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Iptables []*IptableRuleSet `json:"iptables,omitempty"`

	// Port ranges for any servers running in inband LinuxServer clouds. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	KniAllowedServerPorts []*KniPortRange `json:"kni_allowed_server_ports,omitempty"`

	// Number of L7 connections that can be cached per core. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	L7ConnsPerCore *uint32 `json:"l7_conns_per_core,omitempty"`

	// Number of reserved L7 listener connections per core. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	L7ResvdListenConnsPerCore *uint32 `json:"l7_resvd_listen_conns_per_core,omitempty"`

	// Labels associated with this SE group. Field introduced in 20.1.1. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Labels []*KeyValue `json:"labels,omitempty"`

	// Number of requests to dispatch from the request. queue at a regular interval. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LbactionNumRequestsToDispatch *uint32 `json:"lbaction_num_requests_to_dispatch,omitempty"`

	// Maximum retries per request in the request queue. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LbactionRqPerRequestMaxRetries *uint32 `json:"lbaction_rq_per_request_max_retries,omitempty"`

	// Select core with least load for new flow. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LeastLoadCoreSelection *bool `json:"least_load_core_selection,omitempty"`

	// Specifies the license tier which would be used. This field by default inherits the value from cloud. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS, ENTERPRISE_WITH_CLOUD_SERVICES. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseTier *string `json:"license_tier,omitempty"`

	// If no license type is specified then default license enforcement for the cloud type is chosen. Enum options - LIC_BACKEND_SERVERS, LIC_SOCKETS, LIC_CORES, LIC_HOSTS, LIC_SE_BANDWIDTH, LIC_METERED_SE_BANDWIDTH. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseType *string `json:"license_type,omitempty"`

	// Flag to indicate if log files are compressed upon full on the Service Engine. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentCompressLogs *bool `json:"log_agent_compress_logs,omitempty"`

	// Enable debug logs by default on Service Engine. This includes all other debugging logs. Debug logs can also be explcitly enabled from the CLI shell. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentDebugEnabled *bool `json:"log_agent_debug_enabled,omitempty"`

	// Maximum application log file size before rollover. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentFileSzAppl *uint32 `json:"log_agent_file_sz_appl,omitempty"`

	// Maximum connection log file size before rollover. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentFileSzConn *uint32 `json:"log_agent_file_sz_conn,omitempty"`

	// Maximum debug log file size before rollover. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentFileSzDebug *uint32 `json:"log_agent_file_sz_debug,omitempty"`

	// Maximum event log file size before rollover. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentFileSzEvent *uint32 `json:"log_agent_file_sz_event,omitempty"`

	// Minimum storage allocated for logs irrespective of memory and cores. Field introduced in 21.1.1. Unit is MB. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentLogStorageMinSz *uint32 `json:"log_agent_log_storage_min_sz,omitempty"`

	// Maximum concurrent rsync requests initiated from log-agent to the Controller. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentMaxConcurrentRsync *uint32 `json:"log_agent_max_concurrent_rsync,omitempty"`

	// Excess percentage threshold of disk size to trigger cleanup of logs on the Service Engine. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentMaxStorageExcessPercent *uint32 `json:"log_agent_max_storage_excess_percent,omitempty"`

	// Maximum storage on the disk not allocated for logs on the Service Engine. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentMaxStorageIgnorePercent *float32 `json:"log_agent_max_storage_ignore_percent,omitempty"`

	// Minimum storage allocated to any given VirtualService on the Service Engine. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentMinStoragePerVs *uint32 `json:"log_agent_min_storage_per_vs,omitempty"`

	// Internal timer to stall log-agent and prevent it from hogging CPU cycles on the Service Engine. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentSleepInterval *uint32 `json:"log_agent_sleep_interval,omitempty"`

	// Enable trace logs by default on Service Engine. Configuration operations are logged along with other important logs by Service Engine. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentTraceEnabled *bool `json:"log_agent_trace_enabled,omitempty"`

	// Timeout to purge unknown Virtual Service logs from the Service Engine. Field introduced in 21.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogAgentUnknownVsTimer *uint32 `json:"log_agent_unknown_vs_timer,omitempty"`

	// Maximum disk capacity (in MB) to be allocated to an SE. This is exclusively used for debug and log data. Unit is MB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogDisksz *uint32 `json:"log_disksz,omitempty"`

	// SE will log memory allocation related failure to the se_trace file, wherever available. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	LogMallocFailure *bool `json:"log_malloc_failure,omitempty"`

	// Maximum number of file names in a log message. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LogMessageMaxFileListSize *uint32 `json:"log_message_max_file_list_size,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.7. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Maximum number of external health monitors that can run concurrently in a service engine. This helps control the CPU and memory use by external health monitors. Special values are 0- Value will be internally calculated based on cpu and memory. Field introduced in 18.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxConcurrentExternalHm *uint32 `json:"max_concurrent_external_hm,omitempty"`

	// When CPU usage on an SE exceeds this threshold, Virtual Services hosted on this SE may be rebalanced to other SEs to reduce load. A new SE may be created as part of this process. Allowed values are 40-90. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxCPUUsage *int32 `json:"max_cpu_usage,omitempty"`

	// Max bytes that can be allocated in a single mempool. Field introduced in 18.1.5. Unit is MB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxMemoryPerMempool *uint32 `json:"max_memory_per_mempool,omitempty"`

	// Maximum number of HTTP session that will be created. Each session uses about 1kB in the key-value storage in shared memory. Setting this value too high can lead to exhaustion of shared memory and affect services. Allowed values are 1-2000000. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxNumHTTPSessionsToStore *uint32 `json:"max_num_http_sessions_to_store,omitempty"`

	// Configures the maximum number of se_dp processes that handles traffic. If not configured, defaults to the number of CPUs on the SE. If decreased, it will only take effect after SE reboot. Allowed values are 1-128. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	MaxNumSeDps *uint32 `json:"max_num_se_dps,omitempty"`

	// Applicable to Azure platform only. Maximum number of public IPs per Azure LB. . Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxPublicIpsPerLb *uint32 `json:"max_public_ips_per_lb,omitempty"`

	// Maximum number of queues per vnic Setting to '0' utilises all queues that are distributed across dispatcher cores. Allowed values are 0,1,2,4,8,16. Field introduced in 18.2.7, 20.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1), Basic edition(Allowed values- 1), Enterprise with Cloud Services edition.
	MaxQueuesPerVnic *uint32 `json:"max_queues_per_vnic,omitempty"`

	// Applicable to Azure platform only. Maximum number of rules per Azure LB. . Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxRulesPerLb *uint32 `json:"max_rules_per_lb,omitempty"`

	// Maximum number of active Service Engines for the Virtual Service. Allowed values are 1-64. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxScaleoutPerVs *uint32 `json:"max_scaleout_per_vs,omitempty"`

	// Maximum number of Services Engines in this group. Allowed values are 0-1000. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxSe *int32 `json:"max_se,omitempty"`

	// Maximum of number of 4 KB pages allocated to the Linux kernel GRO subsystem for packet coalescing. This parameter is limited to supported kernels only. Requires SE Reboot. Allowed values are 1-17. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxSkbFrags *uint32 `json:"max_skb_frags,omitempty"`

	// Maximum number of Virtual Services that can be placed on a single Service Engine. Allowed values are 1-1000. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxVsPerSe *int32 `json:"max_vs_per_se,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MemReserve *bool `json:"mem_reserve,omitempty"`

	// Indicates the percent of memory reserved for config updates. Allowed values are 0-100. Field introduced in 18.1.2. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MemoryForConfigUpdate *uint32 `json:"memory_for_config_update,omitempty"`

	// Amount of memory for each of the Service Engine virtual machines. Changes to this setting do not affect existing SEs. Allowed values are 2048-262144. Unit is MB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MemoryPerSe *int32 `json:"memory_per_se,omitempty"`

	// Metrics collection mode, 0 = Pull mode.  se_agent pulls metrics from se_dp,  1 = Push mode. se_dp pushes metrics to se_agent.  9 = special value to reset collection state in push mode. . Allowed values are 0-1. Special values are 9- Reset metrics collection state. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MetricsCollectionMode *uint32 `json:"metrics_collection_mode,omitempty"`

	// Management network to use for Avi Service Engines. It is a reference to an object of type Network. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MgmtNetworkRef *string `json:"mgmt_network_ref,omitempty"`

	// Management subnet to use for Avi Service Engines. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MgmtSubnet *IPAddrPrefix `json:"mgmt_subnet,omitempty"`

	// When CPU usage on an SE falls below the minimum threshold, Virtual Services hosted on the SE may be consolidated onto other underutilized SEs. After consolidation, unused Service Engines may then be eligible for deletion. . Allowed values are 20-60. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinCPUUsage *int32 `json:"min_cpu_usage,omitempty"`

	// Minimum number of active Service Engines for the Virtual Service. Allowed values are 1-64. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinScaleoutPerVs *int32 `json:"min_scaleout_per_vs,omitempty"`

	// Minimum number of Services Engines in this group (relevant for SE AutoRebalance only). Allowed values are 0-1000. Field introduced in 17.2.13,18.1.3,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinSe *int32 `json:"min_se,omitempty"`

	// Indicates the percent of memory reserved for connections. Allowed values are 0-100. Field introduced in 18.1.2. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MinimumConnectionMemory *uint32 `json:"minimum_connection_memory,omitempty"`

	// This knob enables the Service Engine to process multicast traffic(For VMware Hypervisor). Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MulticastEnable *bool `json:"multicast_enable,omitempty"`

	// Number of threads to use for log streaming. Allowed values are 1-100. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NLogStreamingThreads *uint32 `json:"n_log_streaming_threads,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Number of threads to poll for netlink messages excluding the thread for default namespace. Requires SE Reboot. Allowed values are 1-32. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NetlinkPollerThreads *uint32 `json:"netlink_poller_threads,omitempty"`

	// Socket buffer size for the netlink sockets. Requires SE Reboot. Allowed values are 1-128. Field introduced in 21.1.1. Unit is MEGA_BYTES. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NetlinkSockBufSize *uint32 `json:"netlink_sock_buf_size,omitempty"`

	// Free the connection stack. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NgxFreeConnectionStack *bool `json:"ngx_free_connection_stack,omitempty"`

	// This setting limits the number of non-significant logs generated per second per core on this SE. Default is 100 logs per second. Set it to zero (0) to deactivate throttling. Field introduced in 17.1.3. Unit is PER_SECOND. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NonSignificantLogThrottle *uint32 `json:"non_significant_log_throttle,omitempty"`

	// Dequeue interval for receive queue from NS HELPER. Allowed values are 1-1000. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	NsHelperDeqIntervalMsec *uint32 `json:"ns_helper_deq_interval_msec,omitempty"`

	// Toggle SE NTP synchronization failure events generation. Disabled by default. Field introduced in 22.1.2. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	NtpSyncFailEvent *bool `json:"ntp_sync_fail_event,omitempty"`

	// Configures the interval at which SE synchronization status with NTP server(s) is verified. A value of zero disables SE NTP synchronization status validation. Allowed values are 120-900. Special values are 0- disable. Field introduced in 22.1.2. Unit is SEC. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	NtpSyncStatusInterval *uint32 `json:"ntp_sync_status_interval,omitempty"`

	// Number of dispatcher cores (0,1,2,4,8 or 16). If set to 0, then number of dispatcher cores is deduced automatically.Requires SE Reboot. Allowed values are 0,1,2,4,8,16. Field introduced in 17.2.12, 18.1.3, 18.2.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	NumDispatcherCores *uint32 `json:"num_dispatcher_cores,omitempty"`

	// Number of queues to each dispatcher. Allowed values are 1-2. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NumDispatcherQueues *uint32 `json:"num_dispatcher_queues,omitempty"`

	// Number of changes in num flow cores sum to ignore. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumFlowCoresSumChangesToIgnore *uint32 `json:"num_flow_cores_sum_changes_to_ignore,omitempty"`

	// Configuration knobs for InterSE Object Distribution. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjsyncConfig *ObjSyncConfig `json:"objsync_config,omitempty"`

	// TCP port on SE management interface for InterSE Object Distribution. Supported only for externally managed security groups. Not supported on full access deployments. Requires SE reboot. Allowed values are 1024-65535. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjsyncPort *uint32 `json:"objsync_port,omitempty"`

	//  Field introduced in 17.1.1. Maximum of 5 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OpenstackAvailabilityZones []string `json:"openstack_availability_zones,omitempty"`

	// Avi Management network name. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OpenstackMgmtNetworkName *string `json:"openstack_mgmt_network_name,omitempty"`

	// Management network UUID. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OpenstackMgmtNetworkUUID *string `json:"openstack_mgmt_network_uuid,omitempty"`

	// Amount of extra memory to be reserved for use by the Operating System on a Service Engine. Unit is MB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsReservedMemory *uint32 `json:"os_reserved_memory,omitempty"`

	// Enable Path MTU Discovery feature for IPv4. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PathMtuDiscoveryV4 *bool `json:"path_mtu_discovery_v4,omitempty"`

	// Enable Path MTU Discovery feature for IPv6. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PathMtuDiscoveryV6 *bool `json:"path_mtu_discovery_v6,omitempty"`

	// Determines the PCAP transmit mode of operation. Requires SE Reboot. Enum options - PCAP_TX_AUTO, PCAP_TX_SOCKET, PCAP_TX_RING. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PcapTxMode *string `json:"pcap_tx_mode,omitempty"`

	// In PCAP mode, reserve a configured portion of TX ring resources for itself and the remaining portion for the RX ring to achieve better balance in terms of queue depth. Requires SE Reboot. Allowed values are 10-100. Field introduced in 20.1.3. Unit is PERCENT. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PcapTxRingRdBalancingFactor *uint32 `json:"pcap_tx_ring_rd_balancing_factor,omitempty"`

	// Per-app SE mode is designed for deploying dedicated load balancers per app (VS). In this mode, each SE is limited to a max of 2 VSs. vCPUs in per-app SEs count towards licensing usage at 25% rate. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	PerApp *bool `json:"per_app,omitempty"`

	// Enable/Disable per VS level admission control.Enabling this feature will cause the connection and packet throttling on a particular VS that has high packet buffer consumption. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PerVsAdmissionControl *bool `json:"per_vs_admission_control,omitempty"`

	// If placement mode is 'Auto', Virtual Services are automatically placed on Service Engines. Enum options - PLACEMENT_MODE_AUTO. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PlacementMode *string `json:"placement_mode,omitempty"`

	// Enable or deactivate real time SE metrics. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RealtimeSeMetrics *MetricsRealTimeUpdate `json:"realtime_se_metrics,omitempty"`

	// Reboot the VM or host on kernel panic. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RebootOnPanic *bool `json:"reboot_on_panic,omitempty"`

	// Routes in VRF are replayed at the specified interval. This should be increased if there are large number of routes. Allowed values are 0-3000. Field introduced in 22.1.3. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ReplayVrfRoutesInterval *uint32 `json:"replay_vrf_routes_interval,omitempty"`

	// Time interval to re-sync SE's time with wall clock time. Allowed values are 8-600000. Field introduced in 20.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResyncTimeInterval *uint32 `json:"resync_time_interval,omitempty"`

	// SDB pipeline flush interval. Allowed values are 1-10000. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SdbFlushInterval *uint32 `json:"sdb_flush_interval,omitempty"`

	// SDB pipeline size. Allowed values are 1-10000. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SdbPipelineSize *uint32 `json:"sdb_pipeline_size,omitempty"`

	// SDB scan count. Allowed values are 1-1000. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SdbScanCount *uint32 `json:"sdb_scan_count,omitempty"`

	// Select the SE bandwidth for the bandwidth license. Enum options - SE_BANDWIDTH_UNLIMITED, SE_BANDWIDTH_25M, SE_BANDWIDTH_200M, SE_BANDWIDTH_1000M, SE_BANDWIDTH_10000M. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- SE_BANDWIDTH_UNLIMITED), Basic edition(Allowed values- SE_BANDWIDTH_UNLIMITED), Enterprise with Cloud Services edition.
	SeBandwidthType *string `json:"se_bandwidth_type,omitempty"`

	// Use to cap the size of debug ring min(se_debug_trace_sz, num_dispatcher_cores). Only applicable to > 8G systems.  Requires SE Reboot. Allowed values are 1,2,4,8,255. Field introduced in 22.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDebugTraceSz *uint32 `json:"se_debug_trace_sz,omitempty"`

	// Delay the cleanup of flowtable entry. To be used under surveillance of Avi Support. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	SeDelayedFlowDelete *bool `json:"se_delayed_flow_delete,omitempty"`

	// Duration to preserve unused Service Engine virtual machines before deleting them. If traffic to a Virtual Service were to spike up abruptly, this SE would still be available to be utilized again rather than creating a new SE. If this value is set to 0, Controller will never delete any SEs and administrator has to manually cleanup unused SEs. Allowed values are 0-525600. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDeprovisionDelay *int32 `json:"se_deprovision_delay,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDosProfile *DosThresholdProfile `json:"se_dos_profile,omitempty"`

	// Internal only. Used to simulate SE - SE HB failure. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDpHmDrops *int32 `json:"se_dp_hm_drops,omitempty"`

	// Number of jiffies between polling interface state. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SeDpIfStatePollInterval *uint32 `json:"se_dp_if_state_poll_interval,omitempty"`

	// Toggle support to run SE datapath instances in isolation on exclusive CPUs. This improves latency and performance. However, this could reduce the total number of se_dp instances created on that SE instance. Supported for >= 8 CPUs. Requires SE reboot. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDpIsolation *bool `json:"se_dp_isolation,omitempty"`

	// Number of CPUs for non se-dp tasks in SE datapath isolation mode. Translates Total cpus minus 'num_non_dp_cpus' for datapath use. It is recommended to reserve an even number of CPUs for hyper-threaded processors. Requires SE reboot. Allowed values are 1-8. Special values are 0- auto. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDpIsolationNumNonDpCpus *uint32 `json:"se_dp_isolation_num_non_dp_cpus,omitempty"`

	// Internal buffer full indicator on the Service Engine beyond which the unfiltered logs are abandoned. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDpLogNfEnqueuePercent *uint32 `json:"se_dp_log_nf_enqueue_percent,omitempty"`

	// Internal buffer full indicator on the Service Engine beyond which the user filtered logs are abandoned. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDpLogUdfEnqueuePercent *uint32 `json:"se_dp_log_udf_enqueue_percent,omitempty"`

	// The highest supported SE-SE Heartbeat protocol version. This version is reported by Secondary SE to Primary SE in Heartbeat response messages. Allowed values are 1-3. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpMaxHbVersion *uint32 `json:"se_dp_max_hb_version,omitempty"`

	// Time (in seconds) service engine waits for after generating a Vnic transmit queue stall event before resetting theNIC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpVnicQueueStallEventSleep *uint32 `json:"se_dp_vnic_queue_stall_event_sleep,omitempty"`

	// Number of consecutive transmit failures to look for before generating a Vnic transmit queue stall event. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpVnicQueueStallThreshold *uint32 `json:"se_dp_vnic_queue_stall_threshold,omitempty"`

	// Time (in milliseconds) to wait for network/NIC recovery on detecting a transmit queue stall after which service engine resets the NIC. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpVnicQueueStallTimeout *uint32 `json:"se_dp_vnic_queue_stall_timeout,omitempty"`

	// Number of consecutive transmit queue stall events in se_dp_vnic_stall_se_restart_window to look for before restarting SE. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpVnicRestartOnQueueStallCount *uint32 `json:"se_dp_vnic_restart_on_queue_stall_count,omitempty"`

	// Window of time (in seconds) during which se_dp_vnic_restart_on_queue_stall_count number of consecutive stalls results in a SE restart. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpVnicStallSeRestartWindow *uint32 `json:"se_dp_vnic_stall_se_restart_window,omitempty"`

	// Determines if DPDK pool mode driver should be used or not   0  Automatically determine based on hypervisor/NIC type 1  Unconditionally use DPDK poll mode driver 2  Don't use DPDK poll mode driver.Requires SE Reboot. Allowed values are 0-2. Field introduced in 18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDpdkPmd *uint32 `json:"se_dpdk_pmd,omitempty"`

	// Enable core dump on assert. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SeDumpCoreOnAssert *bool `json:"se_dump_core_on_assert,omitempty"`

	// Use this to emulate more/less cpus than is actually available. One datapath process is started for each core. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	SeEmulatedCores *uint32 `json:"se_emulated_cores,omitempty"`

	// Flow probe retry count if no replies are received.Requires SE Reboot. Allowed values are 0-5. Field introduced in 18.1.4, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeFlowProbeRetries *uint32 `json:"se_flow_probe_retries,omitempty"`

	// Timeout in milliseconds for flow probe retries.Requires SE Reboot. Allowed values are 20-50. Field introduced in 18.2.5. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeFlowProbeRetryTimer *uint32 `json:"se_flow_probe_retry_timer,omitempty"`

	// Analytics Policy for ServiceEngineGroup. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeGroupAnalyticsPolicy *SeGroupAnalyticsPolicy `json:"se_group_analytics_policy,omitempty"`

	// Controls the distribution of SE data path processes on CPUs which support hyper-threading. Requires hyper-threading to be enabled at host level. Requires SE Reboot. For more details please refer to SE placement KB. Enum options - SE_CPU_HT_AUTO, SE_CPU_HT_SPARSE_DISPATCHER_PRIORITY, SE_CPU_HT_SPARSE_PROXY_PRIORITY, SE_CPU_HT_PACKED_CORES. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHyperthreadedMode *string `json:"se_hyperthreaded_mode,omitempty"`

	// Determines if SE-SE IPC messages are encapsulated in an IP header       0        Automatically determine based on hypervisor type    1        Use IP encap unconditionally    ~[0,1]   Don't use IP encapRequires SE Reboot. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeIPEncapIpc *uint32 `json:"se_ip_encap_ipc,omitempty"`

	// This knob controls the resource availability and burst size used between SE datapath and KNI. This helps in minimising packet drops when there is higher KNI traffic (non-VIP traffic from and to Linux). The factor takes the following values      0-default.     1-doubles the burst size and KNI resources.     2-quadruples the burst size and KNI resources. Allowed values are 0-2. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeKniBurstFactor *uint32 `json:"se_kni_burst_factor,omitempty"`

	// Determines if SE-SE IPC messages use SE interface IP instead of VIP        0        Automatically determine based on hypervisor type    1        Use SE interface IP unconditionally    ~[0,1]   Don't use SE interface IPRequires SE Reboot. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeL3EncapIpc *uint32 `json:"se_l3_encap_ipc,omitempty"`

	// Internal flag that blocks dataplane until all application logs are flushed to log-agent process. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeLogBufferAppBlockingDequeue *bool `json:"se_log_buffer_app_blocking_dequeue,omitempty"`

	// Internal flag that blocks dataplane until all connection logs are flushed to log-agent process. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeLogBufferConnBlockingDequeue *bool `json:"se_log_buffer_conn_blocking_dequeue,omitempty"`

	// Internal flag that blocks dataplane until all outstanding events are flushed to log-agent process. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeLogBufferEventsBlockingDequeue *bool `json:"se_log_buffer_events_blocking_dequeue,omitempty"`

	// Enable or disable Large Receive Optimization for vnics.Supported on VMXnet3.Requires SE Reboot. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeLro *bool `json:"se_lro,omitempty"`

	// The retry count for the multi-producer enqueue before yielding the CPU. To be used under surveillance of Avi Support. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 500), Basic edition(Allowed values- 500), Enterprise with Cloud Services edition.
	SeMpRingRetryCount *uint32 `json:"se_mp_ring_retry_count,omitempty"`

	// MTU for the VNICs of SEs in the SE group. Allowed values are 512-9000. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMtu *uint32 `json:"se_mtu,omitempty"`

	// Prefix to use for virtual machine name of Service Engines. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeNamePrefix *string `json:"se_name_prefix,omitempty"`

	// Internal use only. Used to artificially reduce the available number of packet buffers. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SePacketBufferMax *uint32 `json:"se_packet_buffer_max,omitempty"`

	// Enables lookahead mode of packet receive in PCAP mode. Introduced to overcome an issue with hv_netvsc driver. Lookahead mode attempts to ensure that application and kernel's view of the receive rings are consistent. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePcapLookahead *bool `json:"se_pcap_lookahead,omitempty"`

	// Max number of packets the pcap interface can hold and if the value is 0 the optimum value will be chosen. The optimum value will be chosen based on SE-memory, Cloud Type and Number of Interfaces.Requires SE Reboot. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePcapPktCount *uint32 `json:"se_pcap_pkt_count,omitempty"`

	// Max size of each packet in the pcap interface. Requires SE Reboot. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePcapPktSz *uint32 `json:"se_pcap_pkt_sz,omitempty"`

	// Bypass the kernel's traffic control layer, to deliver packets directly to the driver. Enabling this feature results in egress packets not being captured in host tcpdump. Note   brief packet reordering or loss may occur upon toggle. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePcapQdiscBypass *bool `json:"se_pcap_qdisc_bypass,omitempty"`

	// Frequency in seconds at which periodically a PCAP reinit check is triggered. May be used in conjunction with the configuration pcap_reinit_threshold. (Valid range   15 mins - 12 hours, 0 - disables). Allowed values are 900-43200. Special values are 0- disable. Field introduced in 17.2.13, 18.1.3, 18.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePcapReinitFrequency *uint32 `json:"se_pcap_reinit_frequency,omitempty"`

	// Threshold for input packet receive errors in PCAP mode exceeding which a PCAP reinit is triggered. If not set, an unconditional reinit is performed. This value is checked every pcap_reinit_frequency interval. Field introduced in 17.2.13, 18.1.3, 18.2.1. Unit is METRIC_COUNT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePcapReinitThreshold *uint32 `json:"se_pcap_reinit_threshold,omitempty"`

	// TCP port on SE where echo service will be run. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeProbePort *uint32 `json:"se_probe_port,omitempty"`

	// Rate limiter properties. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRlProp *RateLimiterProperties `json:"se_rl_prop,omitempty"`

	// Minimum time to wait on server between taking sampleswhen sampling the navigation timing data from the end user client. Field introduced in 18.2.6. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRumSamplingNavInterval *uint32 `json:"se_rum_sampling_nav_interval,omitempty"`

	// Percentage of navigation timing data from the end user client, used for sampling to get client insights. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRumSamplingNavPercent *uint32 `json:"se_rum_sampling_nav_percent,omitempty"`

	// Minimum time to wait on server between taking sampleswhen sampling the resource timing data from the end user client. Field introduced in 18.2.6. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRumSamplingResInterval *uint32 `json:"se_rum_sampling_res_interval,omitempty"`

	// Percentage of resource timing data from the end user client used for sampling to get client insight. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRumSamplingResPercent *uint32 `json:"se_rum_sampling_res_percent,omitempty"`

	// Sideband traffic will be handled by a dedicated core.Requires SE Reboot. Field introduced in 16.5.2, 17.1.9, 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeSbDedicatedCore *bool `json:"se_sb_dedicated_core,omitempty"`

	// Number of Sideband threads per SE.Requires SE Reboot. Allowed values are 1-128. Field introduced in 16.5.2, 17.1.9, 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeSbThreads *uint32 `json:"se_sb_threads,omitempty"`

	// Multiplier for SE threads based on vCPU. Allowed values are 1-10. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 1), Basic edition(Allowed values- 1), Enterprise with Cloud Services edition.
	SeThreadMultiplier *uint32 `json:"se_thread_multiplier,omitempty"`

	// Time Tracker Properties for latency audit. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SeTimeTrackerProps *SETimeTrackerProperties `json:"se_time_tracker_props,omitempty"`

	// Traceroute port range. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeTracertPortRange *PortRange `json:"se_tracert_port_range,omitempty"`

	// Determines if Direct Secondary Return (DSR) from secondary SE is active or not  0  Automatically determine based on hypervisor type. 1  Enable tunnel mode - DSR is unconditionally disabled. 2  Disable tunnel mode - DSR is unconditionally enabled. Tunnel mode can be enabled or disabled at run-time. Allowed values are 0-2. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	SeTunnelMode *uint32 `json:"se_tunnel_mode,omitempty"`

	// UDP Port for tunneled packets from secondary to primary SE in Docker bridge mode.Requires SE Reboot. Field introduced in 17.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeTunnelUDPPort *uint32 `json:"se_tunnel_udp_port,omitempty"`

	// Number of packets to batch for transmit to the nic. Requires SE Reboot. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeTxBatchSize *uint32 `json:"se_tx_batch_size,omitempty"`

	// Once the TX queue of the dispatcher reaches this threshold, hardware queues are not polled for further packets. To be used under surveillance of Avi Support. Allowed values are 512-32768. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 2048), Basic edition(Allowed values- 2048), Enterprise with Cloud Services edition.
	SeTxqThreshold *uint32 `json:"se_txq_threshold,omitempty"`

	// Determines if SE-SE IPC messages are encapsulated in a UDP header  0  Automatically determine based on hypervisor type. 1  Use UDP encap unconditionally.Requires SE Reboot. Allowed values are 0-1. Field introduced in 17.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeUDPEncapIpc *uint32 `json:"se_udp_encap_ipc,omitempty"`

	// Determines if DPDK library should be used or not   0  Automatically determine based on hypervisor type 1  Use DPDK if PCAP is not enabled 2  Don't use DPDK. Allowed values are 0-2. Field introduced in 18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeUseDpdk *uint32 `json:"se_use_dpdk,omitempty"`

	// Configure the frequency in milliseconds of software transmit spillover queue flush when enabled. This is necessary to flush any packets in the spillover queue in the absence of a packet transmit in the normal course of operation. Allowed values are 50-500. Special values are 0- disable. Field introduced in 20.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVnicTxSwQueueFlushFrequency *uint32 `json:"se_vnic_tx_sw_queue_flush_frequency,omitempty"`

	// Configure the size of software transmit spillover queue when enabled. Requires SE Reboot. Allowed values are 128-2048. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVnicTxSwQueueSize *uint32 `json:"se_vnic_tx_sw_queue_size,omitempty"`

	// Maximum number of aggregated vs heartbeat packets to send in a batch. Allowed values are 1-256. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVsHbMaxPktsInBatch *uint32 `json:"se_vs_hb_max_pkts_in_batch,omitempty"`

	// Maximum number of virtualservices for which heartbeat messages are aggregated in one packet. Allowed values are 1-1024. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVsHbMaxVsInPkt *uint32 `json:"se_vs_hb_max_vs_in_pkt,omitempty"`

	// Enable SEs to elect a primary amongst themselves in the absence of a connectivity to controller. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	SelfSeElection *bool `json:"self_se_election,omitempty"`

	// Timeout for sending SE_READY without NS HELPER registration completion. Allowed values are 10-600. Field introduced in 21.1.1. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SendSeReadyTimeout *uint32 `json:"send_se_ready_timeout,omitempty"`

	// IPv6 Subnets assigned to the SE group. Required for VS group placement. Field introduced in 18.1.1. Maximum of 128 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceIp6Subnets []*IPAddrPrefix `json:"service_ip6_subnets,omitempty"`

	// Subnets assigned to the SE group. Required for VS group placement. Field introduced in 17.1.1. Maximum of 128 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServiceIPSubnets []*IPAddrPrefix `json:"service_ip_subnets,omitempty"`

	// Minimum required shared memory to apply any configuration. Allowed values are 0-100. Field introduced in 18.1.2. Unit is MB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ShmMinimumConfigMemory *uint32 `json:"shm_minimum_config_memory,omitempty"`

	// This setting limits the number of significant logs generated per second per core on this SE. Default is 100 logs per second. Set it to zero (0) to deactivate throttling. Field introduced in 17.1.3. Unit is PER_SECOND. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SignificantLogThrottle *uint32 `json:"significant_log_throttle,omitempty"`

	// (Beta) Preprocess SSL Client Hello for SNI hostname extension.If set to True, this will apply SNI child's SSL protocol(s), if they are different from SNI Parent's allowed SSL protocol(s). Field introduced in 17.2.12, 18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslPreprocessSniHostname *bool `json:"ssl_preprocess_sni_hostname,omitempty"`

	// Number of SSL sessions that can be cached per VS. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SslSessCachePerVs *uint32 `json:"ssl_sess_cache_per_vs,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// The threshold for the transient shared config memory in the SE. Allowed values are 0-100. Field introduced in 20.1.1. Unit is PERCENT. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TransientSharedMemoryMax *uint32 `json:"transient_shared_memory_max,omitempty"`

	// This setting limits the number of UDF logs generated per second per core on this SE. UDF logs are generated due to the configured client log filters or the rules with logging enabled. Default is 100 logs per second. Set it to zero (0) to deactivate throttling. Field introduced in 17.1.3. Unit is PER_SECOND. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UdfLogThrottle *uint32 `json:"udf_log_throttle,omitempty"`

	// Timeout for backend connection. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpstreamConnectTimeout *uint32 `json:"upstream_connect_timeout,omitempty"`

	// Enable upstream connection pool,. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpstreamConnpoolEnable *bool `json:"upstream_connpool_enable,omitempty"`

	// Timeout for data to be received from backend. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpstreamReadTimeout *uint32 `json:"upstream_read_timeout,omitempty"`

	// Timeout for upstream to become writable. Field introduced in 21.1.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 3600000), Basic edition(Allowed values- 3600000), Enterprise with Cloud Services edition.
	UpstreamSendTimeout *uint32 `json:"upstream_send_timeout,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// If enabled, the datapath CPU utilization is consulted by the auto scale-out logic. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	UseDpUtilForScaleout *bool `json:"use_dp_util_for_scaleout,omitempty"`

	// Enables the use of hyper-threaded cores on SE. Requires SE Reboot. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseHyperthreadedCores *bool `json:"use_hyperthreaded_cores,omitempty"`

	// Enable legacy model of netlink notifications. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UseLegacyNetlink *bool `json:"use_legacy_netlink,omitempty"`

	// Enable InterSE Objsyc distribution framework. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	UseObjsync *bool `json:"use_objsync,omitempty"`

	// Use Standard SKU Azure Load Balancer. By default cloud level flag is set. If not set, it inherits/uses the use_standard_alb flag from the cloud. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseStandardAlb *bool `json:"use_standard_alb,omitempty"`

	// Configuration for User-Agent Cache used in Bot Management. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UserAgentCacheConfig *UserAgentCacheConfig `json:"user_agent_cache_config,omitempty"`

	// Defines in seconds how long before an unused user-defined-metric is garbage collected. Field introduced in 21.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UserDefinedMetricAge *uint32 `json:"user_defined_metric_age,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterClusters *VcenterClusters `json:"vcenter_clusters,omitempty"`

	//  Enum options - VCENTER_DATASTORE_ANY, VCENTER_DATASTORE_LOCAL, VCENTER_DATASTORE_SHARED. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterDatastoreMode *string `json:"vcenter_datastore_mode,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterDatastores []*VcenterDatastore `json:"vcenter_datastores,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterDatastoresInclude *bool `json:"vcenter_datastores_include,omitempty"`

	// Folder to place all the Service Engine virtual machines in vCenter. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterFolder *string `json:"vcenter_folder,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterHosts *VcenterHosts `json:"vcenter_hosts,omitempty"`

	// Parking port group to be used by 9 vnics at the time of SE creation. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterParkingVnicPg *string `json:"vcenter_parking_vnic_pg,omitempty"`

	// VCenter information for scoping at Host/Cluster level. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Vcenters []*PlacementScopeConfig `json:"vcenters,omitempty"`

	// Number of vcpus for each of the Service Engine virtual machines. Changes to this setting do not affect existing SEs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcpusPerSe *int32 `json:"vcpus_per_se,omitempty"`

	// When vip_asg is set, Vip configuration will be managed by Avi.User will be able to configure vip_asg or Vips individually at the time of create. Field introduced in 17.2.12, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipAsg *VipAutoscaleGroup `json:"vip_asg,omitempty"`

	// DHCP ip check interval. Allowed values are 1-1000. Field introduced in 21.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	VnicDhcpIPCheckInterval *uint32 `json:"vnic_dhcp_ip_check_interval,omitempty"`

	// DHCP ip max retries. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	VnicDhcpIPMaxRetries *uint32 `json:"vnic_dhcp_ip_max_retries,omitempty"`

	// wait interval before deleting IP. . Field introduced in 21.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	VnicIPDeleteInterval *uint32 `json:"vnic_ip_delete_interval,omitempty"`

	// Probe vnic interval. Field introduced in 21.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	VnicProbeInterval *uint32 `json:"vnic_probe_interval,omitempty"`

	// Time interval for retrying the failed VNIC RPC requests. Field introduced in 21.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	VnicRPCRetryInterval *uint32 `json:"vnic_rpc_retry_interval,omitempty"`

	// Size of vnicdb command history. Allowed values are 0-65535. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	VnicdbCmdHistorySize *uint32 `json:"vnicdb_cmd_history_size,omitempty"`

	// Ensure primary and secondary Service Engines are deployed on different physical hosts. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition. Special default for Essentials edition is true, Basic edition is true, Enterprise is True.
	VsHostRedundancy *bool `json:"vs_host_redundancy,omitempty"`

	// Time to wait for the scaled in SE to drain existing flows before marking the scalein done. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsScaleinTimeout *uint32 `json:"vs_scalein_timeout,omitempty"`

	// During SE upgrade, Time to wait for the scaled-in SE to drain existing flows before marking the scalein done. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsScaleinTimeoutForUpgrade *uint32 `json:"vs_scalein_timeout_for_upgrade,omitempty"`

	// Time to wait for the scaled out SE to become ready before marking the scaleout done. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsScaleoutTimeout *uint32 `json:"vs_scaleout_timeout,omitempty"`

	// Wait time for primary switchover ready notification after flows are completed. In certain deployments, there may be an additional delay to accept traffic. For example, for BGP, some time is needed for route advertisement. Allowed values are 0-300. Field introduced in 30.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsSePrimarySwitchoverAdditionalWaitTime *uint32 `json:"vs_se_primary_switchover_additional_wait_time,omitempty"`

	// Wait time for sending scalein ready notification after flows are completed. In certain deployments, there may be an additional delay to accept traffic. For example, for BGP, some time is needed for route advertisement. Allowed values are 0-300. Field introduced in 30.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsSeScaleinAdditionalWaitTime *uint32 `json:"vs_se_scalein_additional_wait_time,omitempty"`

	// Wait time for sending scaleout ready notification after Virtual Service is marked UP. In certain deployments, there may be an additional delay to accept traffic. For example, for BGP, some time is needed for route advertisement. Allowed values are 0-300. Field introduced in 18.1.5,18.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsSeScaleoutAdditionalWaitTime *uint32 `json:"vs_se_scaleout_additional_wait_time,omitempty"`

	// Timeout in seconds for Service Engine to sendScaleout Ready notification of a Virtual Service. Allowed values are 0-90. Field introduced in 18.1.5,18.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsSeScaleoutReadyTimeout *uint32 `json:"vs_se_scaleout_ready_timeout,omitempty"`

	// During SE upgrade in a legacy active/standby segroup, Time to wait for the new primary SE to accept flows before marking the switchover done. Field introduced in 17.2.13,18.1.4,18.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsSwitchoverTimeout *uint32 `json:"vs_switchover_timeout,omitempty"`

	// Parameters to place Virtual Services on only a subset of the cores of an SE. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VssPlacement *VssPlacement `json:"vss_placement,omitempty"`

	// If set, Virtual Services will be placed on only a subset of the cores of an SE. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VssPlacementEnabled *bool `json:"vss_placement_enabled,omitempty"`

	// Enable memory pool for WAF.Requires SE Reboot. Field introduced in 17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WafMempool *bool `json:"waf_mempool,omitempty"`

	// Memory pool size used for WAF.Requires SE Reboot. Field introduced in 17.2.3. Unit is KB. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WafMempoolSize *uint32 `json:"waf_mempool_size,omitempty"`
}
