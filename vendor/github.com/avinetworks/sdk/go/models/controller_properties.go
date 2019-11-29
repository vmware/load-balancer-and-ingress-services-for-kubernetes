package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ControllerProperties controller properties
// swagger:model ControllerProperties
type ControllerProperties struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Field introduced in 17.1.1.
	AllowIPForwarding *bool `json:"allow_ip_forwarding,omitempty"`

	// Allow unauthenticated access for special APIs.
	AllowUnauthenticatedApis *bool `json:"allow_unauthenticated_apis,omitempty"`

	// Placeholder for description of property allow_unauthenticated_nodes of obj type ControllerProperties field type str  type boolean
	AllowUnauthenticatedNodes *bool `json:"allow_unauthenticated_nodes,omitempty"`

	//  Allowed values are 0-1440.
	APIIDLETimeout *int32 `json:"api_idle_timeout,omitempty"`

	// Threshold to log request timing in portal_performance.log and Server-Timing response header. Any stage taking longer than 1% of the threshold will be included in the Server-Timing header. Field introduced in 18.1.4, 18.2.1.
	APIPerfLoggingThreshold *int32 `json:"api_perf_logging_threshold,omitempty"`

	// Export configuration in appviewx compatibility mode. Field introduced in 17.1.1.
	AppviewxCompatMode *bool `json:"appviewx_compat_mode,omitempty"`

	// Number of attach_ip_retry_interval.
	AttachIPRetryInterval *int32 `json:"attach_ip_retry_interval,omitempty"`

	// Number of attach_ip_retry_limit.
	AttachIPRetryLimit *int32 `json:"attach_ip_retry_limit,omitempty"`

	// Use Ansible for SE creation in baremetal. Field introduced in 17.2.2.
	BmUseAnsible *bool `json:"bm_use_ansible,omitempty"`

	// Period for auth token cleanup job. Field introduced in 18.1.1.
	CleanupExpiredAuthtokenTimeoutPeriod *int32 `json:"cleanup_expired_authtoken_timeout_period,omitempty"`

	// Period for sessions cleanup job. Field introduced in 18.1.1.
	CleanupSessionsTimeoutPeriod *int32 `json:"cleanup_sessions_timeout_period,omitempty"`

	// Enable/Disable periodic reconcile for all the clouds. Field introduced in 17.2.14,18.1.5,18.2.1.
	CloudReconcile *bool `json:"cloud_reconcile,omitempty"`

	// Period for cluster ip gratuitous arp job.
	ClusterIPGratuitousArpPeriod *int32 `json:"cluster_ip_gratuitous_arp_period,omitempty"`

	// Period for consistency check job. Field introduced in 18.1.1.
	ConsistencyCheckTimeoutPeriod *int32 `json:"consistency_check_timeout_period,omitempty"`

	// Number of crashed_se_reboot.
	CrashedSeReboot *int32 `json:"crashed_se_reboot,omitempty"`

	// Number of dead_se_detection_timer.
	DeadSeDetectionTimer *int32 `json:"dead_se_detection_timer,omitempty"`

	// Period for refresh pool and gslb DNS job.
	DNSRefreshPeriod *int32 `json:"dns_refresh_period,omitempty"`

	// Number of dummy.
	Dummy *int32 `json:"dummy,omitempty"`

	// This setting enables the controller leader to shard API requests to the followers (if any). Field introduced in 18.1.5, 18.2.1.
	EnableAPISharding *bool `json:"enable_api_sharding,omitempty"`

	// Enable/Disable Memory Balancer. Field introduced in 17.2.8.
	EnableMemoryBalancer *bool `json:"enable_memory_balancer,omitempty"`

	// Number of fatal_error_lease_time.
	FatalErrorLeaseTime *int32 `json:"fatal_error_lease_time,omitempty"`

	// Number of max_dead_se_in_grp.
	MaxDeadSeInGrp *int32 `json:"max_dead_se_in_grp,omitempty"`

	// Maximum number of pcap files stored per tenant.
	MaxPcapPerTenant *int32 `json:"max_pcap_per_tenant,omitempty"`

	// Maximum number of consecutive attach IP failures that halts VS placement. Field introduced in 17.2.2.
	MaxSeqAttachIPFailures *int32 `json:"max_seq_attach_ip_failures,omitempty"`

	// Number of max_seq_vnic_failures.
	MaxSeqVnicFailures *int32 `json:"max_seq_vnic_failures,omitempty"`

	// Period for rotate app persistence keys job. Allowed values are 1-1051200. Special values are 0 - 'Disabled'.
	PersistenceKeyRotatePeriod *int32 `json:"persistence_key_rotate_period,omitempty"`

	// Token used for uploading tech-support to portal. Field introduced in 16.4.6,17.1.2.
	PortalToken *string `json:"portal_token,omitempty"`

	// Period for process locked user accounts job. Field introduced in 18.1.1.
	ProcessLockedUseraccountsTimeoutPeriod *int32 `json:"process_locked_useraccounts_timeout_period,omitempty"`

	// Period for process PKI profile job. Field introduced in 18.1.1.
	ProcessPkiProfileTimeoutPeriod *int32 `json:"process_pki_profile_timeout_period,omitempty"`

	// Number of query_host_fail.
	QueryHostFail *int32 `json:"query_host_fail,omitempty"`

	// Version of the safenet package installed on the controller. Field introduced in 16.5.2,17.2.3.
	SafenetHsmVersion *string `json:"safenet_hsm_version,omitempty"`

	// Number of se_create_timeout.
	SeCreateTimeout *int32 `json:"se_create_timeout,omitempty"`

	// Interval between attempting failovers to an SE.
	SeFailoverAttemptInterval *int32 `json:"se_failover_attempt_interval,omitempty"`

	// This setting decides whether SE is to be deployed from the cloud marketplace or to be created by the controller. The setting is applicable only when BYOL license is selected. Enum options - MARKETPLACE, IMAGE. Field introduced in 18.1.4, 18.2.1.
	SeFromMarketplace *string `json:"se_from_marketplace,omitempty"`

	// Number of se_offline_del.
	SeOfflineDel *int32 `json:"se_offline_del,omitempty"`

	// Number of se_vnic_cooldown.
	SeVnicCooldown *int32 `json:"se_vnic_cooldown,omitempty"`

	// Period for secure channel cleanup job.
	SecureChannelCleanupTimeout *int32 `json:"secure_channel_cleanup_timeout,omitempty"`

	// Number of secure_channel_controller_token_timeout.
	SecureChannelControllerTokenTimeout *int32 `json:"secure_channel_controller_token_timeout,omitempty"`

	// Number of secure_channel_se_token_timeout.
	SecureChannelSeTokenTimeout *int32 `json:"secure_channel_se_token_timeout,omitempty"`

	// Pool size used for all fabric commands during se upgrade.
	SeupgradeFabricPoolSize *int32 `json:"seupgrade_fabric_pool_size,omitempty"`

	// Time to wait before marking segroup upgrade as stuck.
	SeupgradeSegroupMinDeadTimeout *int32 `json:"seupgrade_segroup_min_dead_timeout,omitempty"`

	// Number of days for SSL Certificate expiry warning.
	SslCertificateExpiryWarningDays []int64 `json:"ssl_certificate_expiry_warning_days,omitempty,omitempty"`

	// Number of unresponsive_se_reboot.
	UnresponsiveSeReboot *int32 `json:"unresponsive_se_reboot,omitempty"`

	// Time to account for DNS TTL during upgrade. This is in addition to vs_scalein_timeout_for_upgrade in se_group. Field introduced in 17.1.1.
	UpgradeDNSTTL *int32 `json:"upgrade_dns_ttl,omitempty"`

	// Number of upgrade_lease_time.
	UpgradeLeaseTime *int32 `json:"upgrade_lease_time,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Number of vnic_op_fail_time.
	VnicOpFailTime *int32 `json:"vnic_op_fail_time,omitempty"`

	// Time to wait for the scaled out SE to become ready before marking the scaleout done, applies to APIC configuration only.
	VsApicScaleoutTimeout *int32 `json:"vs_apic_scaleout_timeout,omitempty"`

	// Number of vs_awaiting_se_timeout.
	VsAwaitingSeTimeout *int32 `json:"vs_awaiting_se_timeout,omitempty"`

	// Period for rotate VS keys job. Allowed values are 1-1051200. Special values are 0 - 'Disabled'.
	VsKeyRotatePeriod *int32 `json:"vs_key_rotate_period,omitempty"`

	// Time to wait before marking attach IP operation on an SE as failed. Field introduced in 17.2.2.
	VsSeAttachIPFail *int32 `json:"vs_se_attach_ip_fail,omitempty"`

	// Number of vs_se_bootup_fail.
	VsSeBootupFail *int32 `json:"vs_se_bootup_fail,omitempty"`

	// Number of vs_se_create_fail.
	VsSeCreateFail *int32 `json:"vs_se_create_fail,omitempty"`

	// Number of vs_se_ping_fail.
	VsSePingFail *int32 `json:"vs_se_ping_fail,omitempty"`

	// Number of vs_se_vnic_fail.
	VsSeVnicFail *int32 `json:"vs_se_vnic_fail,omitempty"`

	// Number of vs_se_vnic_ip_fail.
	VsSeVnicIPFail *int32 `json:"vs_se_vnic_ip_fail,omitempty"`

	// Number of warmstart_se_reconnect_wait_time.
	WarmstartSeReconnectWaitTime *int32 `json:"warmstart_se_reconnect_wait_time,omitempty"`

	// Timeout for warmstart VS resync. Field introduced in 18.1.4, 18.2.1.
	WarmstartVsResyncWaitTime *int32 `json:"warmstart_vs_resync_wait_time,omitempty"`
}
