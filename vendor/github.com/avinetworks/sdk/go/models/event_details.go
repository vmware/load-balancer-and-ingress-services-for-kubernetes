package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// EventDetails event details
// swagger:model EventDetails
type EventDetails struct {

	// Placeholder for description of property add_networks_details of obj type EventDetails field type str  type object
	AddNetworksDetails *RmAddNetworksEventDetails `json:"add_networks_details,omitempty"`

	// Placeholder for description of property all_seupgrade_event_details of obj type EventDetails field type str  type object
	AllSeupgradeEventDetails *AllSeUpgradeEventDetails `json:"all_seupgrade_event_details,omitempty"`

	// Placeholder for description of property anomaly_details of obj type EventDetails field type str  type object
	AnomalyDetails *AnomalyEventDetails `json:"anomaly_details,omitempty"`

	// Placeholder for description of property apic_agent_bd_vrf_details of obj type EventDetails field type str  type object
	ApicAgentBdVrfDetails *ApicAgentBridgeDomainVrfChange `json:"apic_agent_bd_vrf_details,omitempty"`

	// Placeholder for description of property apic_agent_generic_details of obj type EventDetails field type str  type object
	ApicAgentGenericDetails *ApicAgentGenericEventDetails `json:"apic_agent_generic_details,omitempty"`

	// Placeholder for description of property apic_agent_vs_network_error of obj type EventDetails field type str  type object
	ApicAgentVsNetworkError *ApicAgentVsNetworkError `json:"apic_agent_vs_network_error,omitempty"`

	// Placeholder for description of property avg_uptime_change_details of obj type EventDetails field type str  type object
	AvgUptimeChangeDetails *AvgUptimeChangeDetails `json:"avg_uptime_change_details,omitempty"`

	//  Field introduced in 17.2.10,18.1.2.
	AwsAsgDeletionDetails *AWSASGDelete `json:"aws_asg_deletion_details,omitempty"`

	// Placeholder for description of property aws_asg_notif_details of obj type EventDetails field type str  type object
	AwsAsgNotifDetails *AWSASGNotifDetails `json:"aws_asg_notif_details,omitempty"`

	// Placeholder for description of property aws_infra_details of obj type EventDetails field type str  type object
	AwsInfraDetails *AWSSetup `json:"aws_infra_details,omitempty"`

	// Placeholder for description of property azure_info of obj type EventDetails field type str  type object
	AzureInfo *AzureSetup `json:"azure_info,omitempty"`

	// Azure marketplace license term acceptance event. Field introduced in 18.2.2.
	AzureMpInfo *AzureMarketplace `json:"azure_mp_info,omitempty"`

	// Placeholder for description of property bind_vs_se_details of obj type EventDetails field type str  type object
	BindVsSeDetails *RmBindVsSeEventDetails `json:"bind_vs_se_details,omitempty"`

	// Placeholder for description of property bm_infra_details of obj type EventDetails field type str  type object
	BmInfraDetails *BMSetup `json:"bm_infra_details,omitempty"`

	// Placeholder for description of property bootup_fail_details of obj type EventDetails field type str  type object
	BootupFailDetails *RmSeBootupFailEventDetails `json:"bootup_fail_details,omitempty"`

	// Placeholder for description of property burst_checkout_details of obj type EventDetails field type str  type object
	BurstCheckoutDetails *BurstLicenseDetails `json:"burst_checkout_details,omitempty"`

	// Placeholder for description of property cc_cluster_vip_details of obj type EventDetails field type str  type object
	CcClusterVipDetails *CloudClusterVip `json:"cc_cluster_vip_details,omitempty"`

	// Placeholder for description of property cc_dns_update_details of obj type EventDetails field type str  type object
	CcDNSUpdateDetails *CloudDNSUpdate `json:"cc_dns_update_details,omitempty"`

	// Placeholder for description of property cc_health_details of obj type EventDetails field type str  type object
	CcHealthDetails *CloudHealth `json:"cc_health_details,omitempty"`

	// Placeholder for description of property cc_infra_details of obj type EventDetails field type str  type object
	CcInfraDetails *CloudGeneric `json:"cc_infra_details,omitempty"`

	// Placeholder for description of property cc_ip_details of obj type EventDetails field type str  type object
	CcIPDetails *CloudIPChange `json:"cc_ip_details,omitempty"`

	// Placeholder for description of property cc_parkintf_details of obj type EventDetails field type str  type object
	CcParkintfDetails *CloudVipParkingIntf `json:"cc_parkintf_details,omitempty"`

	// Placeholder for description of property cc_se_vm_details of obj type EventDetails field type str  type object
	CcSeVMDetails *CloudSeVMChange `json:"cc_se_vm_details,omitempty"`

	// Placeholder for description of property cc_sync_services_details of obj type EventDetails field type str  type object
	CcSyncServicesDetails *CloudSyncServices `json:"cc_sync_services_details,omitempty"`

	// Placeholder for description of property cc_tenant_del_details of obj type EventDetails field type str  type object
	CcTenantDelDetails *CloudTenantsDeleted `json:"cc_tenant_del_details,omitempty"`

	// Placeholder for description of property cc_vip_update_details of obj type EventDetails field type str  type object
	CcVipUpdateDetails *CloudVipUpdate `json:"cc_vip_update_details,omitempty"`

	// Placeholder for description of property cc_vnic_details of obj type EventDetails field type str  type object
	CcVnicDetails *CloudVnicChange `json:"cc_vnic_details,omitempty"`

	// Placeholder for description of property cluster_config_failed_details of obj type EventDetails field type str  type object
	ClusterConfigFailedDetails *ClusterConfigFailedEvent `json:"cluster_config_failed_details,omitempty"`

	// Placeholder for description of property cluster_leader_failover_details of obj type EventDetails field type str  type object
	ClusterLeaderFailoverDetails *ClusterLeaderFailoverEvent `json:"cluster_leader_failover_details,omitempty"`

	// Placeholder for description of property cluster_node_add_details of obj type EventDetails field type str  type object
	ClusterNodeAddDetails *ClusterNodeAddEvent `json:"cluster_node_add_details,omitempty"`

	// Placeholder for description of property cluster_node_db_failed_details of obj type EventDetails field type str  type object
	ClusterNodeDbFailedDetails *ClusterNodeDbFailedEvent `json:"cluster_node_db_failed_details,omitempty"`

	// Placeholder for description of property cluster_node_remove_details of obj type EventDetails field type str  type object
	ClusterNodeRemoveDetails *ClusterNodeRemoveEvent `json:"cluster_node_remove_details,omitempty"`

	// Placeholder for description of property cluster_node_shutdown_details of obj type EventDetails field type str  type object
	ClusterNodeShutdownDetails *ClusterNodeShutdownEvent `json:"cluster_node_shutdown_details,omitempty"`

	// Placeholder for description of property cluster_node_started_details of obj type EventDetails field type str  type object
	ClusterNodeStartedDetails *ClusterNodeStartedEvent `json:"cluster_node_started_details,omitempty"`

	// Placeholder for description of property cluster_service_critical_failure_details of obj type EventDetails field type str  type object
	ClusterServiceCriticalFailureDetails *ClusterServiceCriticalFailureEvent `json:"cluster_service_critical_failure_details,omitempty"`

	// Placeholder for description of property cluster_service_failed_details of obj type EventDetails field type str  type object
	ClusterServiceFailedDetails *ClusterServiceFailedEvent `json:"cluster_service_failed_details,omitempty"`

	// Placeholder for description of property cluster_service_restored_details of obj type EventDetails field type str  type object
	ClusterServiceRestoredDetails *ClusterServiceRestoredEvent `json:"cluster_service_restored_details,omitempty"`

	// Placeholder for description of property cluster_warm_reboot_details of obj type EventDetails field type str  type object
	ClusterWarmRebootDetails ClusterWarmRebootEvent `json:"cluster_warm_reboot_details,omitempty"`

	// Placeholder for description of property cntlr_host_list_details of obj type EventDetails field type str  type object
	CntlrHostListDetails *VinfraCntlrHostUnreachableList `json:"cntlr_host_list_details,omitempty"`

	// Placeholder for description of property config_action_details of obj type EventDetails field type str  type object
	ConfigActionDetails *ConfigActionDetails `json:"config_action_details,omitempty"`

	// Placeholder for description of property config_create_details of obj type EventDetails field type str  type object
	ConfigCreateDetails *ConfigCreateDetails `json:"config_create_details,omitempty"`

	// Placeholder for description of property config_delete_details of obj type EventDetails field type str  type object
	ConfigDeleteDetails *ConfigDeleteDetails `json:"config_delete_details,omitempty"`

	// Placeholder for description of property config_password_change_request_details of obj type EventDetails field type str  type object
	ConfigPasswordChangeRequestDetails *ConfigUserPasswordChangeRequest `json:"config_password_change_request_details,omitempty"`

	// Placeholder for description of property config_se_grp_flv_update_details of obj type EventDetails field type str  type object
	ConfigSeGrpFlvUpdateDetails *ConfigSeGrpFlvUpdate `json:"config_se_grp_flv_update_details,omitempty"`

	// Placeholder for description of property config_update_details of obj type EventDetails field type str  type object
	ConfigUpdateDetails *ConfigUpdateDetails `json:"config_update_details,omitempty"`

	// Placeholder for description of property config_user_authrz_rule_details of obj type EventDetails field type str  type object
	ConfigUserAuthrzRuleDetails *ConfigUserAuthrzByRule `json:"config_user_authrz_rule_details,omitempty"`

	// Placeholder for description of property config_user_login_details of obj type EventDetails field type str  type object
	ConfigUserLoginDetails *ConfigUserLogin `json:"config_user_login_details,omitempty"`

	// Placeholder for description of property config_user_logout_details of obj type EventDetails field type str  type object
	ConfigUserLogoutDetails *ConfigUserLogout `json:"config_user_logout_details,omitempty"`

	// Placeholder for description of property config_user_not_authrz_rule_details of obj type EventDetails field type str  type object
	ConfigUserNotAuthrzRuleDetails *ConfigUserNotAuthrzByRule `json:"config_user_not_authrz_rule_details,omitempty"`

	// Placeholder for description of property container_cloud_setup of obj type EventDetails field type str  type object
	ContainerCloudSetup *ContainerCloudSetup `json:"container_cloud_setup,omitempty"`

	// Placeholder for description of property container_cloud_sevice of obj type EventDetails field type str  type object
	ContainerCloudSevice *ContainerCloudService `json:"container_cloud_sevice,omitempty"`

	// Placeholder for description of property cs_infra_details of obj type EventDetails field type str  type object
	CsInfraDetails *CloudStackSetup `json:"cs_infra_details,omitempty"`

	// Placeholder for description of property delete_se_details of obj type EventDetails field type str  type object
	DeleteSeDetails *RmDeleteSeEventDetails `json:"delete_se_details,omitempty"`

	// Placeholder for description of property disable_se_migrate_details of obj type EventDetails field type str  type object
	DisableSeMigrateDetails *DisableSeMigrateEventDetails `json:"disable_se_migrate_details,omitempty"`

	// Placeholder for description of property disc_summary of obj type EventDetails field type str  type object
	DiscSummary *VinfraDiscSummaryDetails `json:"disc_summary,omitempty"`

	// Placeholder for description of property dns_sync_info of obj type EventDetails field type str  type object
	DNSSyncInfo *DNSVsSyncInfo `json:"dns_sync_info,omitempty"`

	// Placeholder for description of property docker_ucp_details of obj type EventDetails field type str  type object
	DockerUcpDetails *DockerUCPSetup `json:"docker_ucp_details,omitempty"`

	// Placeholder for description of property dos_attack_event_details of obj type EventDetails field type str  type object
	DosAttackEventDetails *DosAttackEventDetails `json:"dos_attack_event_details,omitempty"`

	// Placeholder for description of property gcp_info of obj type EventDetails field type str  type object
	GcpInfo *GCPSetup `json:"gcp_info,omitempty"`

	// Placeholder for description of property glb_info of obj type EventDetails field type str  type object
	GlbInfo *GslbStatus `json:"glb_info,omitempty"`

	// Placeholder for description of property gs_info of obj type EventDetails field type str  type object
	GsInfo *GslbServiceStatus `json:"gs_info,omitempty"`

	// Placeholder for description of property host_unavail_details of obj type EventDetails field type str  type object
	HostUnavailDetails *HostUnavailEventDetails `json:"host_unavail_details,omitempty"`

	// Placeholder for description of property hs_details of obj type EventDetails field type str  type object
	HsDetails *HealthScoreDetails `json:"hs_details,omitempty"`

	// Placeholder for description of property ip_fail_details of obj type EventDetails field type str  type object
	IPFailDetails *RmSeIPFailEventDetails `json:"ip_fail_details,omitempty"`

	// Placeholder for description of property license_details of obj type EventDetails field type str  type object
	LicenseDetails *LicenseDetails `json:"license_details,omitempty"`

	// Placeholder for description of property license_expiry_details of obj type EventDetails field type str  type object
	LicenseExpiryDetails *LicenseExpiryDetails `json:"license_expiry_details,omitempty"`

	// Placeholder for description of property marathon_service_port_conflict_details of obj type EventDetails field type str  type object
	MarathonServicePortConflictDetails *MarathonServicePortConflict `json:"marathon_service_port_conflict_details,omitempty"`

	// Placeholder for description of property memory_balancer_info of obj type EventDetails field type str  type object
	MemoryBalancerInfo *MemoryBalancerInfo `json:"memory_balancer_info,omitempty"`

	// Placeholder for description of property mesos_infra_details of obj type EventDetails field type str  type object
	MesosInfraDetails *MesosSetup `json:"mesos_infra_details,omitempty"`

	// Placeholder for description of property metric_threshold_up_details of obj type EventDetails field type str  type object
	MetricThresholdUpDetails *MetricThresoldUpDetails `json:"metric_threshold_up_details,omitempty"`

	// Placeholder for description of property metrics_db_disk_details of obj type EventDetails field type str  type object
	MetricsDbDiskDetails *MetricsDbDiskEventDetails `json:"metrics_db_disk_details,omitempty"`

	// Placeholder for description of property mgmt_nw_change_details of obj type EventDetails field type str  type object
	MgmtNwChangeDetails *VinfraMgmtNwChangeDetails `json:"mgmt_nw_change_details,omitempty"`

	// Placeholder for description of property modify_networks_details of obj type EventDetails field type str  type object
	ModifyNetworksDetails *RmModifyNetworksEventDetails `json:"modify_networks_details,omitempty"`

	// Placeholder for description of property network_subnet_details of obj type EventDetails field type str  type object
	NetworkSubnetDetails *NetworkSubnetInfo `json:"network_subnet_details,omitempty"`

	// Placeholder for description of property nw_subnet_clash_details of obj type EventDetails field type str  type object
	NwSubnetClashDetails *NetworkSubnetClash `json:"nw_subnet_clash_details,omitempty"`

	// Placeholder for description of property nw_summarized_details of obj type EventDetails field type str  type object
	NwSummarizedDetails *SummarizedInfo `json:"nw_summarized_details,omitempty"`

	// Placeholder for description of property oci_info of obj type EventDetails field type str  type object
	OciInfo *OCISetup `json:"oci_info,omitempty"`

	// Placeholder for description of property os_infra_details of obj type EventDetails field type str  type object
	OsInfraDetails *OpenStackClusterSetup `json:"os_infra_details,omitempty"`

	// Placeholder for description of property os_ip_details of obj type EventDetails field type str  type object
	OsIPDetails *OpenStackIPChange `json:"os_ip_details,omitempty"`

	// Placeholder for description of property os_lbaudit_details of obj type EventDetails field type str  type object
	OsLbauditDetails *OpenStackLbProvAuditCheck `json:"os_lbaudit_details,omitempty"`

	// Placeholder for description of property os_lbplugin_op_details of obj type EventDetails field type str  type object
	OsLbpluginOpDetails *OpenStackLbPluginOp `json:"os_lbplugin_op_details,omitempty"`

	// Placeholder for description of property os_se_vm_details of obj type EventDetails field type str  type object
	OsSeVMDetails *OpenStackSeVMChange `json:"os_se_vm_details,omitempty"`

	// Placeholder for description of property os_sync_services_details of obj type EventDetails field type str  type object
	OsSyncServicesDetails *OpenStackSyncServices `json:"os_sync_services_details,omitempty"`

	// Placeholder for description of property os_vnic_details of obj type EventDetails field type str  type object
	OsVnicDetails *OpenStackVnicChange `json:"os_vnic_details,omitempty"`

	// Placeholder for description of property pool_deployment_failure_info of obj type EventDetails field type str  type object
	PoolDeploymentFailureInfo *PoolDeploymentFailureInfo `json:"pool_deployment_failure_info,omitempty"`

	// Placeholder for description of property pool_deployment_success_info of obj type EventDetails field type str  type object
	PoolDeploymentSuccessInfo *PoolDeploymentSuccessInfo `json:"pool_deployment_success_info,omitempty"`

	// Placeholder for description of property pool_deployment_update_info of obj type EventDetails field type str  type object
	PoolDeploymentUpdateInfo *PoolDeploymentUpdateInfo `json:"pool_deployment_update_info,omitempty"`

	// Placeholder for description of property pool_server_delete_details of obj type EventDetails field type str  type object
	PoolServerDeleteDetails *VinfraPoolServerDeleteDetails `json:"pool_server_delete_details,omitempty"`

	// Placeholder for description of property rebalance_migrate_details of obj type EventDetails field type str  type object
	RebalanceMigrateDetails *RebalanceMigrateEventDetails `json:"rebalance_migrate_details,omitempty"`

	// Placeholder for description of property rebalance_scalein_details of obj type EventDetails field type str  type object
	RebalanceScaleinDetails *RebalanceScaleinEventDetails `json:"rebalance_scalein_details,omitempty"`

	// Placeholder for description of property rebalance_scaleout_details of obj type EventDetails field type str  type object
	RebalanceScaleoutDetails *RebalanceScaleoutEventDetails `json:"rebalance_scaleout_details,omitempty"`

	// Placeholder for description of property reboot_se_details of obj type EventDetails field type str  type object
	RebootSeDetails *RmRebootSeEventDetails `json:"reboot_se_details,omitempty"`

	// Placeholder for description of property scheduler_action_info of obj type EventDetails field type str  type object
	SchedulerActionInfo *SchedulerActionDetails `json:"scheduler_action_info,omitempty"`

	// Placeholder for description of property se_bgp_peer_state_change_details of obj type EventDetails field type str  type object
	SeBgpPeerStateChangeDetails *SeBgpPeerStateChangeDetails `json:"se_bgp_peer_state_change_details,omitempty"`

	// Placeholder for description of property se_details of obj type EventDetails field type str  type object
	SeDetails *SeMgrEventDetails `json:"se_details,omitempty"`

	// Placeholder for description of property se_dupip_event_details of obj type EventDetails field type str  type object
	SeDupipEventDetails *SeDupipEventDetails `json:"se_dupip_event_details,omitempty"`

	// Placeholder for description of property se_gateway_heartbeat_failed_details of obj type EventDetails field type str  type object
	SeGatewayHeartbeatFailedDetails *SeGatewayHeartbeatFailedDetails `json:"se_gateway_heartbeat_failed_details,omitempty"`

	// Placeholder for description of property se_gateway_heartbeat_success_details of obj type EventDetails field type str  type object
	SeGatewayHeartbeatSuccessDetails *SeGatewayHeartbeatSuccessDetails `json:"se_gateway_heartbeat_success_details,omitempty"`

	// Placeholder for description of property se_geo_db_details of obj type EventDetails field type str  type object
	SeGeoDbDetails *SeGeoDbDetails `json:"se_geo_db_details,omitempty"`

	// Placeholder for description of property se_hb_event_details of obj type EventDetails field type str  type object
	SeHbEventDetails *SeHBEventDetails `json:"se_hb_event_details,omitempty"`

	// Placeholder for description of property se_hm_gs_details of obj type EventDetails field type str  type object
	SeHmGsDetails *SeHmEventGSDetails `json:"se_hm_gs_details,omitempty"`

	// Placeholder for description of property se_hm_gsgroup_details of obj type EventDetails field type str  type object
	SeHmGsgroupDetails *SeHmEventGslbPoolDetails `json:"se_hm_gsgroup_details,omitempty"`

	// Placeholder for description of property se_hm_pool_details of obj type EventDetails field type str  type object
	SeHmPoolDetails *SeHmEventPoolDetails `json:"se_hm_pool_details,omitempty"`

	// Placeholder for description of property se_hm_vs_details of obj type EventDetails field type str  type object
	SeHmVsDetails *SeHmEventVsDetails `json:"se_hm_vs_details,omitempty"`

	// Placeholder for description of property se_ip6_dad_failed_event_details of obj type EventDetails field type str  type object
	SeIp6DadFailedEventDetails *SeIp6DadFailedEventDetails `json:"se_ip6_dad_failed_event_details,omitempty"`

	// Placeholder for description of property se_ip_added_event_details of obj type EventDetails field type str  type object
	SeIPAddedEventDetails *SeIPAddedEventDetails `json:"se_ip_added_event_details,omitempty"`

	// Placeholder for description of property se_ip_removed_event_details of obj type EventDetails field type str  type object
	SeIPRemovedEventDetails *SeIPRemovedEventDetails `json:"se_ip_removed_event_details,omitempty"`

	// Placeholder for description of property se_ipfailure_event_details of obj type EventDetails field type str  type object
	SeIpfailureEventDetails *SeIpfailureEventDetails `json:"se_ipfailure_event_details,omitempty"`

	// Placeholder for description of property se_licensed_bandwdith_exceeded_event_details of obj type EventDetails field type str  type object
	SeLicensedBandwdithExceededEventDetails *SeLicensedBandwdithExceededEventDetails `json:"se_licensed_bandwdith_exceeded_event_details,omitempty"`

	// Placeholder for description of property se_persistence_details of obj type EventDetails field type str  type object
	SePersistenceDetails *SePersistenceEventDetails `json:"se_persistence_details,omitempty"`

	// Placeholder for description of property se_pool_lb_details of obj type EventDetails field type str  type object
	SePoolLbDetails *SePoolLbEventDetails `json:"se_pool_lb_details,omitempty"`

	// Placeholder for description of property se_thresh_event_details of obj type EventDetails field type str  type object
	SeThreshEventDetails *SeThreshEventDetails `json:"se_thresh_event_details,omitempty"`

	// Placeholder for description of property se_version_check_details of obj type EventDetails field type str  type object
	SeVersionCheckDetails *SeVersionCheckFailedEvent `json:"se_version_check_details,omitempty"`

	// Placeholder for description of property se_vnic_down_event_details of obj type EventDetails field type str  type object
	SeVnicDownEventDetails *SeVnicDownEventDetails `json:"se_vnic_down_event_details,omitempty"`

	// Placeholder for description of property se_vnic_tx_queue_stall_event_details of obj type EventDetails field type str  type object
	SeVnicTxQueueStallEventDetails *SeVnicTxQueueStallEventDetails `json:"se_vnic_tx_queue_stall_event_details,omitempty"`

	// Placeholder for description of property se_vnic_up_event_details of obj type EventDetails field type str  type object
	SeVnicUpEventDetails *SeVnicUpEventDetails `json:"se_vnic_up_event_details,omitempty"`

	// Placeholder for description of property se_vs_fault_event_details of obj type EventDetails field type str  type object
	SeVsFaultEventDetails *SeVsFaultEventDetails `json:"se_vs_fault_event_details,omitempty"`

	// Placeholder for description of property semigrate_event_details of obj type EventDetails field type str  type object
	SemigrateEventDetails *SeMigrateEventDetails `json:"semigrate_event_details,omitempty"`

	// Placeholder for description of property server_autoscale_failed_info of obj type EventDetails field type str  type object
	ServerAutoscaleFailedInfo *ServerAutoScaleFailedInfo `json:"server_autoscale_failed_info,omitempty"`

	// Placeholder for description of property server_autoscalein_complete_info of obj type EventDetails field type str  type object
	ServerAutoscaleinCompleteInfo *ServerAutoScaleInCompleteInfo `json:"server_autoscalein_complete_info,omitempty"`

	// Placeholder for description of property server_autoscalein_info of obj type EventDetails field type str  type object
	ServerAutoscaleinInfo *ServerAutoScaleInInfo `json:"server_autoscalein_info,omitempty"`

	// Placeholder for description of property server_autoscaleout_complete_info of obj type EventDetails field type str  type object
	ServerAutoscaleoutCompleteInfo *ServerAutoScaleOutCompleteInfo `json:"server_autoscaleout_complete_info,omitempty"`

	// Placeholder for description of property server_autoscaleout_info of obj type EventDetails field type str  type object
	ServerAutoscaleoutInfo *ServerAutoScaleOutInfo `json:"server_autoscaleout_info,omitempty"`

	// Placeholder for description of property seupgrade_disrupted_details of obj type EventDetails field type str  type object
	SeupgradeDisruptedDetails *SeUpgradeVsDisruptedEventDetails `json:"seupgrade_disrupted_details,omitempty"`

	// Placeholder for description of property seupgrade_event_details of obj type EventDetails field type str  type object
	SeupgradeEventDetails *SeUpgradeEventDetails `json:"seupgrade_event_details,omitempty"`

	// Placeholder for description of property seupgrade_migrate_details of obj type EventDetails field type str  type object
	SeupgradeMigrateDetails *SeUpgradeMigrateEventDetails `json:"seupgrade_migrate_details,omitempty"`

	// Placeholder for description of property seupgrade_scalein_details of obj type EventDetails field type str  type object
	SeupgradeScaleinDetails *SeUpgradeScaleinEventDetails `json:"seupgrade_scalein_details,omitempty"`

	// Placeholder for description of property seupgrade_scaleout_details of obj type EventDetails field type str  type object
	SeupgradeScaleoutDetails *SeUpgradeScaleoutEventDetails `json:"seupgrade_scaleout_details,omitempty"`

	// Placeholder for description of property spawn_se_details of obj type EventDetails field type str  type object
	SpawnSeDetails *RmSpawnSeEventDetails `json:"spawn_se_details,omitempty"`

	// Placeholder for description of property ssl_expire_details of obj type EventDetails field type str  type object
	SslExpireDetails *SSLExpireDetails `json:"ssl_expire_details,omitempty"`

	// Placeholder for description of property ssl_export_details of obj type EventDetails field type str  type object
	SslExportDetails *SSLExportDetails `json:"ssl_export_details,omitempty"`

	// Placeholder for description of property ssl_renew_details of obj type EventDetails field type str  type object
	SslRenewDetails *SSLRenewDetails `json:"ssl_renew_details,omitempty"`

	// Placeholder for description of property ssl_renew_failed_details of obj type EventDetails field type str  type object
	SslRenewFailedDetails *SSLRenewFailedDetails `json:"ssl_renew_failed_details,omitempty"`

	// Placeholder for description of property switchover_details of obj type EventDetails field type str  type object
	SwitchoverDetails *SwitchoverEventDetails `json:"switchover_details,omitempty"`

	// Placeholder for description of property switchover_fail_details of obj type EventDetails field type str  type object
	SwitchoverFailDetails *SwitchoverFailEventDetails `json:"switchover_fail_details,omitempty"`

	// Placeholder for description of property system_upgrade_details of obj type EventDetails field type str  type object
	SystemUpgradeDetails *SystemUpgradeDetails `json:"system_upgrade_details,omitempty"`

	// Placeholder for description of property unbind_vs_se_details of obj type EventDetails field type str  type object
	UnbindVsSeDetails *RmUnbindVsSeEventDetails `json:"unbind_vs_se_details,omitempty"`

	// Placeholder for description of property vca_infra_details of obj type EventDetails field type str  type object
	VcaInfraDetails *VCASetup `json:"vca_infra_details,omitempty"`

	// Placeholder for description of property vcenter_connectivity_status of obj type EventDetails field type str  type object
	VcenterConnectivityStatus *VinfraVcenterConnectivityStatus `json:"vcenter_connectivity_status,omitempty"`

	// Placeholder for description of property vcenter_details of obj type EventDetails field type str  type object
	VcenterDetails *VinfraVcenterBadCredentials `json:"vcenter_details,omitempty"`

	// Placeholder for description of property vcenter_disc_failure of obj type EventDetails field type str  type object
	VcenterDiscFailure *VinfraVcenterDiscoveryFailure `json:"vcenter_disc_failure,omitempty"`

	// Placeholder for description of property vcenter_network_limit of obj type EventDetails field type str  type object
	VcenterNetworkLimit *VinfraVcenterNetworkLimit `json:"vcenter_network_limit,omitempty"`

	// Placeholder for description of property vcenter_obj_delete_details of obj type EventDetails field type str  type object
	VcenterObjDeleteDetails *VinfraVcenterObjDeleteDetails `json:"vcenter_obj_delete_details,omitempty"`

	// Placeholder for description of property vip_autoscale of obj type EventDetails field type str  type object
	VipAutoscale *VipScaleDetails `json:"vip_autoscale,omitempty"`

	// Placeholder for description of property vip_dns_info of obj type EventDetails field type str  type object
	VipDNSInfo *DNSRegisterInfo `json:"vip_dns_info,omitempty"`

	// Placeholder for description of property vm_details of obj type EventDetails field type str  type object
	VMDetails *VinfraVMDetails `json:"vm_details,omitempty"`

	// Placeholder for description of property vs_awaitingse_details of obj type EventDetails field type str  type object
	VsAwaitingseDetails *VsAwaitingSeEventDetails `json:"vs_awaitingse_details,omitempty"`

	// Placeholder for description of property vs_error_details of obj type EventDetails field type str  type object
	VsErrorDetails *VsErrorEventDetails `json:"vs_error_details,omitempty"`

	// Placeholder for description of property vs_fsm_details of obj type EventDetails field type str  type object
	VsFsmDetails *VsFsmEventDetails `json:"vs_fsm_details,omitempty"`

	// Placeholder for description of property vs_initialplacement_details of obj type EventDetails field type str  type object
	VsInitialplacementDetails *VsInitialPlacementEventDetails `json:"vs_initialplacement_details,omitempty"`

	// Placeholder for description of property vs_migrate_details of obj type EventDetails field type str  type object
	VsMigrateDetails *VsMigrateEventDetails `json:"vs_migrate_details,omitempty"`

	// Placeholder for description of property vs_pool_nw_fltr_details of obj type EventDetails field type str  type object
	VsPoolNwFltrDetails *VsPoolNwFilterEventDetails `json:"vs_pool_nw_fltr_details,omitempty"`

	// Placeholder for description of property vs_scalein_details of obj type EventDetails field type str  type object
	VsScaleinDetails *VsScaleInEventDetails `json:"vs_scalein_details,omitempty"`

	// Placeholder for description of property vs_scaleout_details of obj type EventDetails field type str  type object
	VsScaleoutDetails *VsScaleOutEventDetails `json:"vs_scaleout_details,omitempty"`
}
