package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerProperties controller properties
// swagger:model ControllerProperties
type ControllerProperties struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Allow non-admin tenants to update admin VrfContext and Network objects. Field introduced in 18.2.7, 20.1.1.
	AllowAdminNetworkUpdates *bool `json:"allow_admin_network_updates,omitempty"`

	//  Field introduced in 17.1.1.
	AllowIPForwarding *bool `json:"allow_ip_forwarding,omitempty"`

	// Allow unauthenticated access for special APIs.
	AllowUnauthenticatedApis *bool `json:"allow_unauthenticated_apis,omitempty"`

	// Placeholder for description of property allow_unauthenticated_nodes of obj type ControllerProperties field type str  type boolean
	AllowUnauthenticatedNodes *bool `json:"allow_unauthenticated_nodes,omitempty"`

	//  Allowed values are 0-1440. Unit is MIN.
	APIIDLETimeout *int32 `json:"api_idle_timeout,omitempty"`

	// Threshold to log request timing in portal_performance.log and Server-Timing response header. Any stage taking longer than 1% of the threshold will be included in the Server-Timing header. Field introduced in 18.1.4, 18.2.1. Unit is MILLISECONDS.
	APIPerfLoggingThreshold *int32 `json:"api_perf_logging_threshold,omitempty"`

	// Export configuration in appviewx compatibility mode. Field introduced in 17.1.1. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	AppviewxCompatMode *bool `json:"appviewx_compat_mode,omitempty"`

	// Period for which asynchronous patch requests are queued. Allowed values are 30-120. Special values are 0 - 'Deactivated'. Field introduced in 18.2.11, 20.1.3. Unit is SEC.
	AsyncPatchMergePeriod *int32 `json:"async_patch_merge_period,omitempty"`

	// Duration for which asynchronous patch requests should be kept, after being marked as SUCCESS or FAIL. Allowed values are 5-120. Field introduced in 18.2.11, 20.1.3. Unit is MIN.
	AsyncPatchRequestCleanupDuration *int32 `json:"async_patch_request_cleanup_duration,omitempty"`

	//  Unit is SEC.
	AttachIPRetryInterval *int32 `json:"attach_ip_retry_interval,omitempty"`

	// Number of attach_ip_retry_limit.
	AttachIPRetryLimit *int32 `json:"attach_ip_retry_limit,omitempty"`

	// Use Ansible for SE creation in baremetal. Field introduced in 17.2.2.
	BmUseAnsible *bool `json:"bm_use_ansible,omitempty"`

	// Period for auth token cleanup job. Field introduced in 18.1.1. Unit is MIN.
	CleanupExpiredAuthtokenTimeoutPeriod *int32 `json:"cleanup_expired_authtoken_timeout_period,omitempty"`

	// Period for sessions cleanup job. Field introduced in 18.1.1. Unit is MIN.
	CleanupSessionsTimeoutPeriod *int32 `json:"cleanup_sessions_timeout_period,omitempty"`

	// Enable/Disable periodic reconcile for all the clouds. Field introduced in 17.2.14,18.1.5,18.2.1.
	CloudReconcile *bool `json:"cloud_reconcile,omitempty"`

	// Period for cluster ip gratuitous arp job. Unit is MIN.
	ClusterIPGratuitousArpPeriod *int32 `json:"cluster_ip_gratuitous_arp_period,omitempty"`

	// Period for consistency check job. Field introduced in 18.1.1. Unit is MIN.
	ConsistencyCheckTimeoutPeriod *int32 `json:"consistency_check_timeout_period,omitempty"`

	// Periodically collect stats. Field introduced in 20.1.3. Unit is MIN.
	ControllerResourceInfoCollectionPeriod *int32 `json:"controller_resource_info_collection_period,omitempty"`

	//  Unit is SEC.
	CrashedSeReboot *int32 `json:"crashed_se_reboot,omitempty"`

	//  Unit is SEC.
	DeadSeDetectionTimer *int32 `json:"dead_se_detection_timer,omitempty"`

	// Minimum api timeout value.If this value is not 60, it will be the default timeout for all APIs that do not have a specific timeout.If an API has a specific timeout but is less than this value, this value will become the new timeout. Allowed values are 60-3600. Field introduced in 18.2.6. Unit is SEC.
	DefaultMinimumAPITimeout *int32 `json:"default_minimum_api_timeout,omitempty"`

	// Period for refresh pool and gslb DNS job. Unit is MIN. Allowed in Basic(Allowed values- 60) edition, Essentials(Allowed values- 60) edition, Enterprise edition.
	DNSRefreshPeriod *int32 `json:"dns_refresh_period,omitempty"`

	// Number of dummy.
	Dummy *int32 `json:"dummy,omitempty"`

	// Allow editing of system limits. Keep in mind that these system limits have been carefully selected based on rigorous testing in our testig environments. Modifying these limits could destabilize your cluster. Do this at your own risk!. Field introduced in 20.1.1.
	EditSystemLimits *bool `json:"edit_system_limits,omitempty"`

	// This setting enables the controller leader to shard API requests to the followers (if any). Field introduced in 18.1.5, 18.2.1.
	EnableAPISharding *bool `json:"enable_api_sharding,omitempty"`

	// Enable/Disable Memory Balancer. Field introduced in 17.2.8.
	EnableMemoryBalancer *bool `json:"enable_memory_balancer,omitempty"`

	//  Unit is SEC.
	FatalErrorLeaseTime *int32 `json:"fatal_error_lease_time,omitempty"`

	// Federated datastore will not cleanup diffs unless they are at least this duration in the past. Field introduced in 20.1.1. Unit is HOURS.
	FederatedDatastoreCleanupDuration *int64 `json:"federated_datastore_cleanup_duration,omitempty"`

	// Period for file object cleanup job. Field introduced in 20.1.1. Unit is MIN.
	FileObjectCleanupPeriod *int32 `json:"file_object_cleanup_period,omitempty"`

	// Number of max_dead_se_in_grp.
	MaxDeadSeInGrp *int32 `json:"max_dead_se_in_grp,omitempty"`

	// Maximum number of pcap files stored per tenant.
	MaxPcapPerTenant *int32 `json:"max_pcap_per_tenant,omitempty"`

	// Maximum delay possible to add to se_spawn_retry_interval after successive SE spawn failure. Field introduced in 20.1.1. Unit is SEC.
	MaxSeSpawnIntervalDelay *int32 `json:"max_se_spawn_interval_delay,omitempty"`

	// Maximum number of consecutive attach IP failures that halts VS placement. Field introduced in 17.2.2.
	MaxSeqAttachIPFailures *int32 `json:"max_seq_attach_ip_failures,omitempty"`

	// Number of max_seq_vnic_failures.
	MaxSeqVnicFailures *int32 `json:"max_seq_vnic_failures,omitempty"`

	// Maximum number of threads in threadpool used by cloud connector CCVIPBGWorker. Allowed values are 1-100. Field introduced in 20.1.3.
	MaxThreadsCcVipBgWorker *int32 `json:"max_threads_cc_vip_bg_worker,omitempty"`

	// Network and VrfContext objects from the admin tenant will not be shared to non-admin tenants unless admin permissions are granted. Field introduced in 18.2.7, 20.1.1.
	PermissionScopedSharedAdminNetworks *bool `json:"permission_scoped_shared_admin_networks,omitempty"`

	// Period for rotate app persistence keys job. Allowed values are 1-1051200. Special values are 0 - 'Disabled'. Unit is MIN. Allowed in Basic(Allowed values- 0) edition, Essentials(Allowed values- 0) edition, Enterprise edition.
	PersistenceKeyRotatePeriod *int32 `json:"persistence_key_rotate_period,omitempty"`

	// Burst limit on number of incoming requests0 to disable. Field introduced in 20.1.1.
	PortalRequestBurstLimit *int32 `json:"portal_request_burst_limit,omitempty"`

	// Maximum average number of requests allowed per second0 to disable. Field introduced in 20.1.1. Unit is PER_SECOND.
	PortalRequestRateLimit *int32 `json:"portal_request_rate_limit,omitempty"`

	// Token used for uploading tech-support to portal. Field introduced in 16.4.6,17.1.2.
	PortalToken *string `json:"portal_token,omitempty"`

	// Period for process locked user accounts job. Field introduced in 18.1.1. Unit is MIN.
	ProcessLockedUseraccountsTimeoutPeriod *int32 `json:"process_locked_useraccounts_timeout_period,omitempty"`

	// Period for process PKI profile job. Field introduced in 18.1.1. Unit is MIN.
	ProcessPkiProfileTimeoutPeriod *int32 `json:"process_pki_profile_timeout_period,omitempty"`

	//  Unit is SEC.
	QueryHostFail *int32 `json:"query_host_fail,omitempty"`

	// Version of the safenet package installed on the controller. Field introduced in 16.5.2,17.2.3.
	SafenetHsmVersion *string `json:"safenet_hsm_version,omitempty"`

	//  Unit is SEC.
	SeCreateTimeout *int32 `json:"se_create_timeout,omitempty"`

	// Interval between attempting failovers to an SE. Unit is SEC.
	SeFailoverAttemptInterval *int32 `json:"se_failover_attempt_interval,omitempty"`

	// This setting decides whether SE is to be deployed from the cloud marketplace or to be created by the controller. The setting is applicable only when BYOL license is selected. Enum options - MARKETPLACE, IMAGE. Field introduced in 18.1.4, 18.2.1.
	SeFromMarketplace *string `json:"se_from_marketplace,omitempty"`

	//  Unit is SEC.
	SeOfflineDel *int32 `json:"se_offline_del,omitempty"`

	// Default retry period before attempting another Service Engine spawn in SE Group. Field introduced in 20.1.1. Unit is SEC.
	SeSpawnRetryInterval *int32 `json:"se_spawn_retry_interval,omitempty"`

	//  Unit is SEC.
	SeVnicCooldown *int32 `json:"se_vnic_cooldown,omitempty"`

	// Period for secure channel cleanup job. Unit is MIN.
	SecureChannelCleanupTimeout *int32 `json:"secure_channel_cleanup_timeout,omitempty"`

	//  Unit is MIN.
	SecureChannelControllerTokenTimeout *int32 `json:"secure_channel_controller_token_timeout,omitempty"`

	//  Unit is MIN.
	SecureChannelSeTokenTimeout *int32 `json:"secure_channel_se_token_timeout,omitempty"`

	// This parameter defines the number of simultaneous SE image downloads in a SeGroup. It is used to pace the SE downloads so that controller network/CPU bandwidth is a bounded operation. A value of 0 will disable the pacing scheme and all the SE(s) in the SeGroup will attempt to download the image. . Field introduced in 18.2.6.
	SeupgradeCopyPoolSize *int32 `json:"seupgrade_copy_pool_size,omitempty"`

	// Pool size used for all fabric commands during se upgrade.
	SeupgradeFabricPoolSize *int32 `json:"seupgrade_fabric_pool_size,omitempty"`

	// Time to wait before marking segroup upgrade as stuck. Unit is SEC.
	SeupgradeSegroupMinDeadTimeout *int32 `json:"seupgrade_segroup_min_dead_timeout,omitempty"`

	// SSL Certificates in the admin tenant can be used in non-admin tenants. Field introduced in 18.2.5.
	SharedSslCertificates *bool `json:"shared_ssl_certificates,omitempty"`

	// Number of days for SSL Certificate expiry warning. Unit is DAYS.
	SslCertificateExpiryWarningDays []int64 `json:"ssl_certificate_expiry_warning_days,omitempty,omitempty"`

	//  Unit is SEC.
	UnresponsiveSeReboot *int32 `json:"unresponsive_se_reboot,omitempty"`

	// Time to account for DNS TTL during upgrade. This is in addition to vs_scalein_timeout_for_upgrade in se_group. Field introduced in 17.1.1. Unit is SEC. Allowed in Basic(Allowed values- 5) edition, Essentials(Allowed values- 5) edition, Enterprise edition.
	UpgradeDNSTTL *int32 `json:"upgrade_dns_ttl,omitempty"`

	// Amount of time Controller waits for a large-sized SE (>=128GB memory) to reconnect after it is rebooted during upgrade. Field introduced in 18.2.10, 20.1.1. Unit is SEC.
	UpgradeFatSeLeaseTime *int32 `json:"upgrade_fat_se_lease_time,omitempty"`

	// Amount of time Controller waits for a regular-sized SE (<128GB memory) to reconnect after it is rebooted during upgrade. Starting 18.2.10/20.1.1, the default time has increased from 360 seconds to 600 seconds. Unit is SEC.
	UpgradeLeaseTime *int32 `json:"upgrade_lease_time,omitempty"`

	// This parameter defines the upper-bound value of the VS scale-in or VS scale-out operation executed in the SeScaleIn and SeScale context.  User can tweak this parameter to a higher value if the Segroup gets suspended due to SeScalein or SeScaleOut timeout failure typically associated with high number of VS(es) scaled out. . Field introduced in 18.2.10, 20.1.1. Unit is SEC.
	UpgradeSePerVsScaleOpsTxnTime *int32 `json:"upgrade_se_per_vs_scale_ops_txn_time,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	//  Unit is SEC.
	VnicOpFailTime *int32 `json:"vnic_op_fail_time,omitempty"`

	// Time to wait for the scaled out SE to become ready before marking the scaleout done, applies to APIC configuration only. Unit is SEC.
	VsApicScaleoutTimeout *int32 `json:"vs_apic_scaleout_timeout,omitempty"`

	//  Unit is SEC.
	VsAwaitingSeTimeout *int32 `json:"vs_awaiting_se_timeout,omitempty"`

	// Period for rotate VS keys job. Allowed values are 1-1051200. Special values are 0 - 'Disabled'. Unit is MIN.
	VsKeyRotatePeriod *int32 `json:"vs_key_rotate_period,omitempty"`

	// Interval for checking scaleout_ready status while controller is waiting for ScaleOutReady RPC from the Service Engine. Field introduced in 18.2.2. Unit is SEC.
	VsScaleoutReadyCheckInterval *int32 `json:"vs_scaleout_ready_check_interval,omitempty"`

	// Time to wait before marking attach IP operation on an SE as failed. Field introduced in 17.2.2. Unit is SEC.
	VsSeAttachIPFail *int32 `json:"vs_se_attach_ip_fail,omitempty"`

	//  Unit is SEC.
	VsSeBootupFail *int32 `json:"vs_se_bootup_fail,omitempty"`

	//  Unit is SEC.
	VsSeCreateFail *int32 `json:"vs_se_create_fail,omitempty"`

	//  Unit is SEC.
	VsSePingFail *int32 `json:"vs_se_ping_fail,omitempty"`

	//  Unit is SEC.
	VsSeVnicFail *int32 `json:"vs_se_vnic_fail,omitempty"`

	//  Unit is SEC.
	VsSeVnicIPFail *int32 `json:"vs_se_vnic_ip_fail,omitempty"`

	//  Unit is SEC.
	WarmstartSeReconnectWaitTime *int32 `json:"warmstart_se_reconnect_wait_time,omitempty"`

	// Timeout for warmstart VS resync. Field introduced in 18.1.4, 18.2.1. Unit is SEC.
	WarmstartVsResyncWaitTime *int32 `json:"warmstart_vs_resync_wait_time,omitempty"`
}
