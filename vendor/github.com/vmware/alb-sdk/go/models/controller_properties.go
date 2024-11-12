// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerProperties controller properties
// swagger:model ControllerProperties
type ControllerProperties struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Allow non-admin tenants to update admin VrfContext and Network objects. Field introduced in 18.2.7, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowAdminNetworkUpdates *bool `json:"allow_admin_network_updates,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowIPForwarding *bool `json:"allow_ip_forwarding,omitempty"`

	// Allow unauthenticated access for special APIs. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowUnauthenticatedApis *bool `json:"allow_unauthenticated_apis,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllowUnauthenticatedNodes *bool `json:"allow_unauthenticated_nodes,omitempty"`

	//  Allowed values are 0-1440. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	APIIDLETimeout *uint32 `json:"api_idle_timeout,omitempty"`

	// Threshold to log request timing in portal_performance.log and Server-Timing response header. Any stage taking longer than 1% of the threshold will be included in the Server-Timing header. Field introduced in 18.1.4, 18.2.1. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	APIPerfLoggingThreshold *uint32 `json:"api_perf_logging_threshold,omitempty"`

	// Export configuration in appviewx compatibility mode. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	AppviewxCompatMode *bool `json:"appviewx_compat_mode,omitempty"`

	// Period for which asynchronous patch requests are queued. Allowed values are 30-120. Special values are 0 - Deactivated. Field introduced in 18.2.11, 20.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AsyncPatchMergePeriod *uint32 `json:"async_patch_merge_period,omitempty"`

	// Duration for which asynchronous patch requests should be kept, after being marked as SUCCESS or FAIL. Allowed values are 5-120. Field introduced in 18.2.11, 20.1.3. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AsyncPatchRequestCleanupDuration *uint32 `json:"async_patch_request_cleanup_duration,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AttachIPRetryInterval *uint32 `json:"attach_ip_retry_interval,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AttachIPRetryLimit *uint32 `json:"attach_ip_retry_limit,omitempty"`

	// Use Ansible for SE creation in baremetal. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BmUseAnsible *bool `json:"bm_use_ansible,omitempty"`

	// Enforce VsVip FQDN syntax checks. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	CheckVsvipFqdnSyntax *bool `json:"check_vsvip_fqdn_syntax,omitempty"`

	// Period for auth token cleanup job. Field introduced in 18.1.1. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CleanupExpiredAuthtokenTimeoutPeriod *uint32 `json:"cleanup_expired_authtoken_timeout_period,omitempty"`

	// Period for sessions cleanup job. Field introduced in 18.1.1. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CleanupSessionsTimeoutPeriod *uint32 `json:"cleanup_sessions_timeout_period,omitempty"`

	// Time in minutes to wait between consecutive cloud discovery cycles. Allowed values are 1-1440. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloudDiscoveryInterval *uint32 `json:"cloud_discovery_interval,omitempty"`

	// Enable/Disable periodic reconcile for all the clouds. Field introduced in 17.2.14,18.1.5,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudReconcile *bool `json:"cloud_reconcile,omitempty"`

	// Time in minutes to wait between consecutive cloud reconcile cycles. Allowed values are 1-1440. Field introduced in 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloudReconcileInterval *uint32 `json:"cloud_reconcile_interval,omitempty"`

	// Period for cluster ip gratuitous arp job. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterIPGratuitousArpPeriod *uint32 `json:"cluster_ip_gratuitous_arp_period,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Period for consistency check job. Field introduced in 18.1.1. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConsistencyCheckTimeoutPeriod *uint32 `json:"consistency_check_timeout_period,omitempty"`

	// Periodically collect stats. Field introduced in 20.1.3. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ControllerResourceInfoCollectionPeriod *uint32 `json:"controller_resource_info_collection_period,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CrashedSeReboot *uint32 `json:"crashed_se_reboot,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DeadSeDetectionTimer *uint32 `json:"dead_se_detection_timer,omitempty"`

	// Minimum api timeout value.If this value is not 60, it will be the default timeout for all APIs that do not have a specific timeout.If an API has a specific timeout but is less than this value, this value will become the new timeout. Allowed values are 60-3600. Field introduced in 18.2.6. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DefaultMinimumAPITimeout *uint32 `json:"default_minimum_api_timeout,omitempty"`

	// The amount of time the controller will wait before deleting an offline SE after it has been rebooted. For unresponsive SEs, the total time will be  unresponsive_se_reboot + del_offline_se_after_reboot_delay. For crashed SEs, the total time will be crashed_se_reboot + del_offline_se_after_reboot_delay. Field introduced in 20.1.5. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DelOfflineSeAfterRebootDelay *uint32 `json:"del_offline_se_after_reboot_delay,omitempty"`

	// Amount of time to wait after last Detach IP failure before attempting next Detach IP retry. Field introduced in 21.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DetachIPRetryInterval *uint32 `json:"detach_ip_retry_interval,omitempty"`

	// Maximum number of Detach IP retries. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DetachIPRetryLimit *uint32 `json:"detach_ip_retry_limit,omitempty"`

	// Time to wait before marking Detach IP as failed. Field introduced in 21.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DetachIPTimeout *uint32 `json:"detach_ip_timeout,omitempty"`

	// Period for refresh pool and gslb DNS job. Unit is MIN. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 60), Basic edition(Allowed values- 60), Enterprise with Cloud Services edition.
	DNSRefreshPeriod *uint32 `json:"dns_refresh_period,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Dummy *uint32 `json:"dummy,omitempty"`

	// Allow editing of system limits. Keep in mind that these system limits have been carefully selected based on rigorous testing in our testig environments. Modifying these limits could destabilize your cluster. Do this at your own risk!. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EditSystemLimits *bool `json:"edit_system_limits,omitempty"`

	// This setting enables the controller leader to shard API requests to the followers (if any). Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableAPISharding *bool `json:"enable_api_sharding,omitempty"`

	// Enable/Disable Memory Balancer. Field introduced in 17.2.8. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableMemoryBalancer *bool `json:"enable_memory_balancer,omitempty"`

	// Enable stopping of individual processes if process cross the given threshold limit, even when the total controller memory usage is belowits threshold limit. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnablePerProcessStop *bool `json:"enable_per_process_stop,omitempty"`

	// Enable printing of cached logs inside Resource Manager. Used for debugging purposes only. Field introduced in 20.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableResmgrLogCachePrint *bool `json:"enable_resmgr_log_cache_print,omitempty"`

	// Maximum number of goroutines for event manager process. Allowed values are 1-64. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EventManagerMaxGoroutines *uint32 `json:"event_manager_max_goroutines,omitempty"`

	// Maximum number of subscribers for event manager process. Allowed values are 1-6. Special values are 0 - Disabled. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EventManagerMaxSubscribers *uint32 `json:"event_manager_max_subscribers,omitempty"`

	// Log instances for event manager processing delay; recorded whenever event processing delay exceeds configured interval specified in seconds. Allowed values are 1-5. Special values are 0 - Disabled. Field introduced in 30.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EventManagerProcessingTimeThreshold *uint32 `json:"event_manager_processing_time_threshold,omitempty"`

	// False Positive learning configuration. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FalsePositiveLearningConfig *FalsePositiveLearningConfig `json:"false_positive_learning_config,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FatalErrorLeaseTime *uint32 `json:"fatal_error_lease_time,omitempty"`

	// Federated datastore will not cleanup diffs unless they are at least this duration in the past. Field introduced in 20.1.1. Unit is HOURS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FederatedDatastoreCleanupDuration *uint64 `json:"federated_datastore_cleanup_duration,omitempty"`

	// Period for file object cleanup job. Field introduced in 20.1.1. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FileObjectCleanupPeriod *uint32 `json:"file_object_cleanup_period,omitempty"`

	// This is the max number of file versions that will be retained for a file referenced by the local FileObject. Subsequent uploads of file will result in the file rotation of the older version and the latest version retained. Example  When a file Upload is done for the first time, there will be a v1 version. Subsequent uploads will get mapped to v1, v2 and v3 versions. On the fourth upload of the file, the v1 will be file rotated and v2, v3 and v4 will be retained. Allowed values are 1-5. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FileobjectMaxFileVersions *uint32 `json:"fileobject_max_file_versions,omitempty"`

	// Batch size for the vs_mgr to perform datastrorecleanup during a GSLB purge. Allowed values are 50-1200. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GslbPurgeBatchSize *uint32 `json:"gslb_purge_batch_size,omitempty"`

	// Sleep time in the vs_mgr during a FederatedPurge RPC call. Allowed values are 50-100. Field introduced in 22.1.3. Unit is MILLISECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GslbPurgeSleepTimeMs *uint32 `json:"gslb_purge_sleep_time_ms,omitempty"`

	// Ignore the vrf_context filter for /networksubnetlist API. Field introduced in 22.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IgnoreVrfInNetworksubnetlist *bool `json:"ignore_vrf_in_networksubnetlist,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxDeadSeInGrp *uint32 `json:"max_dead_se_in_grp,omitempty"`

	// Maximum number of pcap files stored per tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxPcapPerTenant *uint32 `json:"max_pcap_per_tenant,omitempty"`

	// Maximum delay possible to add to se_spawn_retry_interval after successive SE spawn failure. Field introduced in 20.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxSeSpawnIntervalDelay *uint32 `json:"max_se_spawn_interval_delay,omitempty"`

	// Maximum number of consecutive attach IP failures that halts VS placement. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxSeqAttachIPFailures *uint32 `json:"max_seq_attach_ip_failures,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxSeqVnicFailures *uint32 `json:"max_seq_vnic_failures,omitempty"`

	// Maximum number of threads in threadpool used by cloud connector CCVIPBGWorker. Allowed values are 1-100. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxThreadsCcVipBgWorker *uint32 `json:"max_threads_cc_vip_bg_worker,omitempty"`

	// Network and VrfContext objects from the admin tenant will not be shared to non-admin tenants unless admin permissions are granted. Field introduced in 18.2.7, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PermissionScopedSharedAdminNetworks *bool `json:"permission_scoped_shared_admin_networks,omitempty"`

	// Period for rotate app persistence keys job. Allowed values are 1-1051200. Special values are 0 - Disabled. Unit is MIN. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 0), Basic edition(Allowed values- 0), Enterprise with Cloud Services edition.
	PersistenceKeyRotatePeriod *uint32 `json:"persistence_key_rotate_period,omitempty"`

	// Burst limit on number of incoming requests. 0 to disable. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PortalRequestBurstLimit *uint32 `json:"portal_request_burst_limit,omitempty"`

	// Maximum average number of requests allowed per second. 0 to disable. Field introduced in 20.1.1. Unit is PER_SECOND. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PortalRequestRateLimit *uint32 `json:"portal_request_rate_limit,omitempty"`

	// Token used for uploading tech-support to portal. Field introduced in 16.4.6,17.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PortalToken *string `json:"portal_token,omitempty"`

	// Period for which postgres vacuum are executed. Allowed values are 30-40320. Special values are 0 - Deactivated. Field introduced in 22.1.3. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PostgresVacuumPeriod *uint32 `json:"postgres_vacuum_period,omitempty"`

	// Period for process locked user accounts job. Field introduced in 18.1.1. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ProcessLockedUseraccountsTimeoutPeriod *uint32 `json:"process_locked_useraccounts_timeout_period,omitempty"`

	// Period for process PKI profile job. Field introduced in 18.1.1. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ProcessPkiProfileTimeoutPeriod *uint32 `json:"process_pki_profile_timeout_period,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	QueryHostFail *uint32 `json:"query_host_fail,omitempty"`

	// Period for each cycle of log caching in Resource Manager. At the end of each cycle, the in memory cached log history will be cleared. Field introduced in 20.1.5. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResmgrLogCachingPeriod *uint32 `json:"resmgr_log_caching_period,omitempty"`

	// Restrict read access to cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RestrictCloudReadAccess *bool `json:"restrict_cloud_read_access,omitempty"`

	// Version of the safenet package installed on the controller. Field introduced in 16.5.2,17.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SafenetHsmVersion *string `json:"safenet_hsm_version,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeCreateTimeout *uint32 `json:"se_create_timeout,omitempty"`

	// Interval between attempting failovers to an SE. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeFailoverAttemptInterval *uint32 `json:"se_failover_attempt_interval,omitempty"`

	// This setting decides whether SE is to be deployed from the cloud marketplace or to be created by the controller. The setting is applicable only when BYOL license is selected. Enum options - MARKETPLACE, IMAGE_SE. Field introduced in 18.1.4, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeFromMarketplace *string `json:"se_from_marketplace,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeOfflineDel *uint32 `json:"se_offline_del,omitempty"`

	// Default retry period before attempting another Service Engine spawn in SE Group. Field introduced in 20.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeSpawnRetryInterval *uint32 `json:"se_spawn_retry_interval,omitempty"`

	// Timeout for flows cleanup by ServiceEngine during Upgrade.Internal knob  to be exercised under the surveillance of VMware AVI support team. Field introduced in 22.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeUpgradeFlowCleanupTimeout *uint32 `json:"se_upgrade_flow_cleanup_timeout,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVnicCooldown *uint32 `json:"se_vnic_cooldown,omitempty"`

	// Duration to wait after last vNIC addition before proceeding with vNIC garbage collection. Used for testing purposes. Field introduced in 20.1.4. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeVnicGcWaitTime *uint32 `json:"se_vnic_gc_wait_time,omitempty"`

	// Period for secure channel cleanup job. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecureChannelCleanupTimeout *uint32 `json:"secure_channel_cleanup_timeout,omitempty"`

	//  Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecureChannelControllerTokenTimeout *uint32 `json:"secure_channel_controller_token_timeout,omitempty"`

	//  Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecureChannelSeTokenTimeout *uint32 `json:"secure_channel_se_token_timeout,omitempty"`

	// This parameter defines the buffer size during SE image downloads in a SeGroup. It is used to pace the SE downloads so that controller network/CPU bandwidth is a bounded operation. Field introduced in 22.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeupgradeCopyBufferSize *uint32 `json:"seupgrade_copy_buffer_size,omitempty"`

	// This parameter defines the number of simultaneous SE image downloads in a SeGroup. It is used to pace the SE downloads so that controller network/CPU bandwidth is a bounded operation. A value of 0 will disable the pacing scheme and all the SE(s) in the SeGroup will attempt to download the image. . Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeupgradeCopyPoolSize *uint32 `json:"seupgrade_copy_pool_size,omitempty"`

	// The pool size is used to control the number of concurrent segroup upgrades. This field value takes affect upon controller warm reboot. Allowed values are 2-20. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeupgradeFabricPoolSize *uint32 `json:"seupgrade_fabric_pool_size,omitempty"`

	// Time to wait before marking segroup upgrade as stuck. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeupgradeSegroupMinDeadTimeout *uint32 `json:"seupgrade_segroup_min_dead_timeout,omitempty"`

	// SSL Certificates in the admin tenant can be used in non-admin tenants. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SharedSslCertificates *bool `json:"shared_ssl_certificates,omitempty"`

	// Time interval (in seconds) between retires for skopeo commands. Field introduced in 30.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SkopeoRetryInterval *uint32 `json:"skopeo_retry_interval,omitempty"`

	// Number of times to try skopeo commands for remote image registries. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SkopeoRetryLimit *uint32 `json:"skopeo_retry_limit,omitempty"`

	// Soft Limit on the minimum SE Memory that an SE needs to have on SE Register. Field introduced in 30.1.1. Unit is MB. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SoftMinMemPerSeLimit *uint32 `json:"soft_min_mem_per_se_limit,omitempty"`

	// Number of days for SSL Certificate expiry warning. Unit is DAYS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslCertificateExpiryWarningDays []int64 `json:"ssl_certificate_expiry_warning_days,omitempty,omitempty"`

	// Time in minutes to wait between cleanup of SystemReports. Allowed values are 15-300. Field introduced in 22.1.6, 30.2.1. Unit is MIN. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SystemReportCleanupInterval *uint32 `json:"system_report_cleanup_interval,omitempty"`

	// Number of SystemReports retained in the system. Once the number of system reports exceed this threshold, the oldest SystemReport will be removed and the latest one retained. i.e. the SystemReport will be rotated and the reports don't exceed the threshold. Allowed values are 5-50. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SystemReportLimit *uint32 `json:"system_report_limit,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UnresponsiveSeReboot *uint32 `json:"unresponsive_se_reboot,omitempty"`

	// Number of times to retry a DNS entry update/delete operation. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpdateDNSEntryRetryLimit *uint32 `json:"update_dns_entry_retry_limit,omitempty"`

	// Timeout period for a DNS entry update/delete operation. Field introduced in 21.1.4. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UpdateDNSEntryTimeout *uint32 `json:"update_dns_entry_timeout,omitempty"`

	// Time to account for DNS TTL during upgrade. This is in addition to vs_scalein_timeout_for_upgrade in se_group. Field introduced in 17.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 5), Basic edition(Allowed values- 5), Enterprise with Cloud Services edition.
	UpgradeDNSTTL *uint32 `json:"upgrade_dns_ttl,omitempty"`

	// Amount of time Controller waits for a large-sized SE (>=128GB memory) to reconnect after it is rebooted during upgrade. Field introduced in 18.2.10, 20.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpgradeFatSeLeaseTime *uint32 `json:"upgrade_fat_se_lease_time,omitempty"`

	// Amount of time Controller waits for a regular-sized SE (<128GB memory) to reconnect after it is rebooted during upgrade. Starting 18.2.10/20.1.1, the default time has increased from 360 seconds to 600 seconds. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpgradeLeaseTime *uint32 `json:"upgrade_lease_time,omitempty"`

	// This parameter defines the upper-bound value of the VS scale-in or VS scale-out operation executed in the SeScaleIn and SeScale context.  User can tweak this parameter to a higher value if the Segroup gets suspended due to SeScalein or SeScaleOut timeout failure typically associated with high number of VS(es) scaled out. . Field introduced in 18.2.10, 20.1.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpgradeSePerVsScaleOpsTxnTime *uint32 `json:"upgrade_se_per_vs_scale_ops_txn_time,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Configuration for User-Agent Cache used in Bot Management. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UserAgentCacheConfig *UserAgentCacheConfig `json:"user_agent_cache_config,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VnicOpFailTime *uint32 `json:"vnic_op_fail_time,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsAwaitingSeTimeout *uint32 `json:"vs_awaiting_se_timeout,omitempty"`

	// Period for rotate VS keys job. Allowed values are 1-1051200. Special values are 0 - Disabled. Unit is MIN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsKeyRotatePeriod *uint32 `json:"vs_key_rotate_period,omitempty"`

	// Interval for checking scaleout_ready status while controller is waiting for ScaleOutReady RPC from the Service Engine. Field introduced in 18.2.2. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsScaleoutReadyCheckInterval *int32 `json:"vs_scaleout_ready_check_interval,omitempty"`

	// Time to wait before marking attach IP operation on an SE as failed. Field introduced in 17.2.2. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsSeAttachIPFail *uint32 `json:"vs_se_attach_ip_fail,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsSeBootupFail *uint32 `json:"vs_se_bootup_fail,omitempty"`

	// Wait for longer for patch SEs to boot up. Field introduced in 30.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsSeBootupFailPatch *uint32 `json:"vs_se_bootup_fail_patch,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsSeCreateFail *uint32 `json:"vs_se_create_fail,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsSePingFail *uint32 `json:"vs_se_ping_fail,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsSeVnicFail *uint32 `json:"vs_se_vnic_fail,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsSeVnicIPFail *uint32 `json:"vs_se_vnic_ip_fail,omitempty"`

	// vSphere HA monitor detection timeout. If vsphere_ha_enabled is true and the controller is not able to reach the SE, placement will wait for this duration for vsphere_ha_inprogress to be marked true before taking corrective action. Field introduced in 20.1.7, 21.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaDetectionTimeout *uint32 `json:"vsphere_ha_detection_timeout,omitempty"`

	// vSphere HA monitor recovery timeout. Once vsphere_ha_inprogress is set to true (meaning host failure detected and vSphere HA will recover the Service Engine), placement will wait for at least this duration for the SE to reconnect to the controller before taking corrective action. Field introduced in 20.1.7, 21.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaRecoveryTimeout *uint32 `json:"vsphere_ha_recovery_timeout,omitempty"`

	// vSphere HA monitor timer interval for sending cc_check_se_status to Cloud Connector. Field introduced in 20.1.7, 21.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsphereHaTimerInterval *uint32 `json:"vsphere_ha_timer_interval,omitempty"`

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WarmstartSeReconnectWaitTime *uint32 `json:"warmstart_se_reconnect_wait_time,omitempty"`

	// Timeout for warmstart VS resync. Field introduced in 18.1.4, 18.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WarmstartVsResyncWaitTime *uint32 `json:"warmstart_vs_resync_wait_time,omitempty"`
}
