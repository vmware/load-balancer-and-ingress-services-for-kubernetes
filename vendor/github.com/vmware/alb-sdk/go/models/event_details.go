// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// EventDetails event details
// swagger:model EventDetails
type EventDetails struct {

	// Adaptive replication event e.g. DNS VS, config version. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AdaptreplEvent *AdaptReplEventInfo `json:"adaptrepl_event,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AddNetworksDetails *RmAddNetworksEventDetails `json:"add_networks_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AlbservicesCaseDetails *ALBServicesCase `json:"albservices_case_details,omitempty"`

	// ALBservices file download event details. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AlbservicesFileDownloadDetails *ALBServicesFileDownload `json:"albservices_file_download_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AlbservicesFileUploadDetails *ALBServicesFileUpload `json:"albservices_file_upload_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AlbservicesStatusDetails *ALBServicesStatusDetails `json:"albservices_status_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllSeupgradeEventDetails *AllSeUpgradeEventDetails `json:"all_seupgrade_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AnomalyDetails *AnomalyEventDetails `json:"anomaly_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	APIVersionDeprecated *APIVersionDeprecated `json:"api_version_deprecated,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppSignatureEventData *AppSignatureEventData `json:"app_signature_event_data,omitempty"`

	//  Field introduced in 22.1.6,30.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AsyncPatchState *AsyncPatchState `json:"async_patch_state,omitempty"`

	// Details for Attach IP status. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AttachIPStatusDetails *AttachIPStatusEventDetails `json:"attach_ip_status_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvgUptimeChangeDetails *AvgUptimeChangeDetails `json:"avg_uptime_change_details,omitempty"`

	//  Field introduced in 17.2.10,18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AwsAsgDeletionDetails *AWSASGDelete `json:"aws_asg_deletion_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AwsAsgNotifDetails *AWSASGNotifDetails `json:"aws_asg_notif_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AwsInfraDetails *AWSSetup `json:"aws_infra_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AzureInfo *AzureSetup `json:"azure_info,omitempty"`

	// Azure marketplace license term acceptance event. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AzureMpInfo *AzureMarketplace `json:"azure_mp_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BindVsSeDetails *RmBindVsSeEventDetails `json:"bind_vs_se_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BmInfraDetails *BMSetup `json:"bm_infra_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BootupFailDetails *RmSeBootupFailEventDetails `json:"bootup_fail_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	BurstCheckoutDetails *BurstLicenseDetails `json:"burst_checkout_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcClusterVipDetails *CloudClusterVip `json:"cc_cluster_vip_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcDNSUpdateDetails *CloudDNSUpdate `json:"cc_dns_update_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcHealthDetails *CloudHealth `json:"cc_health_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcInfraDetails *CloudGeneric `json:"cc_infra_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcIPDetails *CloudIPChange `json:"cc_ip_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcParkintfDetails *CloudVipParkingIntf `json:"cc_parkintf_details,omitempty"`

	//  Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcScalesetNotifDetails *CCScaleSetNotifDetails `json:"cc_scaleset_notif_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcSeVMDetails *CloudSeVMChange `json:"cc_se_vm_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcSyncServicesDetails *CloudSyncServices `json:"cc_sync_services_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcTenantDelDetails *CloudTenantsDeleted `json:"cc_tenant_del_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcVipUpdateDetails *CloudVipUpdate `json:"cc_vip_update_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcVnicDetails *CloudVnicChange `json:"cc_vnic_details,omitempty"`

	// Central license refresh details. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CentralLicenseRefreshDetails *CentralLicenseRefreshDetails `json:"central_license_refresh_details,omitempty"`

	// Central license subscription details. Field introduced in 21.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CentralLicenseSubscriptionDetails *CentralLicenseSubscriptionDetails `json:"central_license_subscription_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudAsgNotifDetails *CloudASGNotifDetails `json:"cloud_asg_notif_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudAutoscalingConfigFailureDetails *CloudAutoscalingConfigFailureDetails `json:"cloud_autoscaling_config_failure_details,omitempty"`

	// Cloud Routes event. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloudRouteNotifDetails *CloudRouteNotifDetails `json:"cloud_route_notif_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterConfigFailedDetails *ClusterConfigFailedEvent `json:"cluster_config_failed_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterLeaderFailoverDetails *ClusterLeaderFailoverEvent `json:"cluster_leader_failover_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterNodeAddDetails *ClusterNodeAddEvent `json:"cluster_node_add_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterNodeDbFailedDetails *ClusterNodeDbFailedEvent `json:"cluster_node_db_failed_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterNodeRemoveDetails *ClusterNodeRemoveEvent `json:"cluster_node_remove_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterNodeShutdownDetails *ClusterNodeShutdownEvent `json:"cluster_node_shutdown_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterNodeStartedDetails *ClusterNodeStartedEvent `json:"cluster_node_started_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterServiceCriticalFailureDetails *ClusterServiceCriticalFailureEvent `json:"cluster_service_critical_failure_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterServiceFailedDetails *ClusterServiceFailedEvent `json:"cluster_service_failed_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClusterServiceRestoredDetails *ClusterServiceRestoredEvent `json:"cluster_service_restored_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ClustifyCheckDetails *ClustifyCheckEvent `json:"clustify_check_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CntlrHostListDetails *VinfraCntlrHostUnreachableList `json:"cntlr_host_list_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigActionDetails *ConfigActionDetails `json:"config_action_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigCreateDetails *ConfigCreateDetails `json:"config_create_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigDeleteDetails *ConfigDeleteDetails `json:"config_delete_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigPasswordChangeRequestDetails *ConfigUserPasswordChangeRequest `json:"config_password_change_request_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigSeGrpFlvUpdateDetails *ConfigSeGrpFlvUpdate `json:"config_se_grp_flv_update_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigUpdateDetails *ConfigUpdateDetails `json:"config_update_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigUserAuthrzRuleDetails *ConfigUserAuthrzByRule `json:"config_user_authrz_rule_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigUserLoginDetails *ConfigUserLogin `json:"config_user_login_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigUserLogoutDetails *ConfigUserLogout `json:"config_user_logout_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigUserNotAuthrzRuleDetails *ConfigUserNotAuthrzByRule `json:"config_user_not_authrz_rule_details,omitempty"`

	// Connection event. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ConnectionEvent *ConnectionEventDetails `json:"connection_event,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ContainerCloudBatchSetup *ContainerCloudBatchSetup `json:"container_cloud_batch_setup,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ContainerCloudSetup *ContainerCloudSetup `json:"container_cloud_setup,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ContainerCloudSevice *ContainerCloudService `json:"container_cloud_sevice,omitempty"`

	//  Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ControllerDiscontinuousTimeChangeEventDetails *ControllerDiscontinuousTimeChangeEventDetails `json:"controller_discontinuous_time_change_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ControllerLicenseReconcileDetails *ControllerLicenseReconcileDetails `json:"controller_license_reconcile_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CrsDeploymentFailure *CRSDeploymentFailure `json:"crs_deployment_failure,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CrsDeploymentSuccess *CRSDeploymentSuccess `json:"crs_deployment_success,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CrsDetails *CRSDetails `json:"crs_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CrsUpdateDetails *CRSUpdateDetails `json:"crs_update_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CsInfraDetails *CloudStackSetup `json:"cs_infra_details,omitempty"`

	// Database error event. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DatabaseEventInfo *DatabaseEventInfo `json:"database_event_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DeleteSeDetails *RmDeleteSeEventDetails `json:"delete_se_details,omitempty"`

	// Details for Detach IP status. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DetachIPStatusDetails *DetachIPStatusEventDetails `json:"detach_ip_status_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableSeMigrateDetails *DisableSeMigrateEventDetails `json:"disable_se_migrate_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DiscSummary *VinfraDiscSummaryDetails `json:"disc_summary,omitempty"`

	// Log files exsiting on controller need to be cleanup. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DiskCleanupEventDetails *LogMgrCleanupEventDetails `json:"disk_cleanup_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSQueryError *DNSQueryError `json:"dns_query_error,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSSyncInfo *DNSVsSyncInfo `json:"dns_sync_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DockerUcpDetails *DockerUCPSetup `json:"docker_ucp_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DosAttackEventDetails *DosAttackEventDetails `json:"dos_attack_event_details,omitempty"`

	// False positive details. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FalsePositiveDetails *FalsePositiveDetails `json:"false_positive_details,omitempty"`

	// File object event. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FileobjectDetails *FileObjectDetails `json:"fileobject_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GcpCloudRouterInfo *GCPCloudRouterUpdate `json:"gcp_cloud_router_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GcpInfo *GCPSetup `json:"gcp_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GenericAuditComplianceEventInfo *AuditComplianceEventInfo `json:"generic_audit_compliance_event_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GlbInfo *GslbStatus `json:"glb_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GsInfo *GslbServiceStatus `json:"gs_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostUnavailDetails *HostUnavailEventDetails `json:"host_unavail_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HsDetails *HealthScoreDetails `json:"hs_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPFailDetails *RmSeIPFailEventDetails `json:"ip_fail_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPThreatDbEventData *IPThreatDBEventData `json:"ip_threat_db_event_data,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseDetails *LicenseDetails `json:"license_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseExpiryDetails *LicenseExpiryDetails `json:"license_expiry_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseTierSwitchDetails *LicenseTierSwitchDetiails `json:"license_tier_switch_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseTransactionDetails *LicenseTransactionDetails `json:"license_transaction_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LogAgentEventDetails *LogAgentEventDetail `json:"log_agent_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MarathonServicePortConflictDetails *MarathonServicePortConflict `json:"marathon_service_port_conflict_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MemoryBalancerInfo *MemoryBalancerInfo `json:"memory_balancer_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MesosInfraDetails *MesosSetup `json:"mesos_infra_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricThresholdUpDetails *MetricThresoldUpDetails `json:"metric_threshold_up_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsDbDiskDetails *MetricsDbDiskEventDetails `json:"metrics_db_disk_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsDbQueueFullDetails *MetricsDbQueueFullEventDetails `json:"metrics_db_queue_full_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsDbQueueHealthyDetails *MetricsDbQueueHealthyEventDetails `json:"metrics_db_queue_healthy_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsDbSyncFailureDetails *MetricsDbSyncFailureEventDetails `json:"metrics_db_sync_failure_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MetricsGrpcAuthFailureDetails *MetricsGRPCAuthFailureDetails `json:"metrics_grpc_auth_failure_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MgmtNwChangeDetails *VinfraMgmtNwChangeDetails `json:"mgmt_nw_change_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ModifyNetworksDetails *RmModifyNetworksEventDetails `json:"modify_networks_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NetworkSubnetDetails *NetworkSubnetInfo `json:"network_subnet_details,omitempty"`

	// NSX-T ServiceInsertion VirtualEndpoint event. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NsxtEndpointDetails *NsxtSIEndpointDetails `json:"nsxt_endpoint_details,omitempty"`

	// Nsxt Image event. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NsxtImgDetails *NsxtImageDetails `json:"nsxt_img_details,omitempty"`

	// Nsxt cloud event. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NsxtInfo *NsxtSetup `json:"nsxt_info,omitempty"`

	// NSX-T ServiceInsertion RedirectPolicy event. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NsxtPolicyDetails *NsxtSIpolicyDetails `json:"nsxt_policy_details,omitempty"`

	// NSX-T ServiceInsertion RedirectRule event. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NsxtRuleDetails *NsxtSIRuleDetails `json:"nsxt_rule_details,omitempty"`

	// NSX-T ServiceInsertion Service event. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NsxtServiceDetails *NsxtSIServiceDetails `json:"nsxt_service_details,omitempty"`

	// NSX-T Tier1(s) Segment(s) event details. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NsxtT1SegDetails *NsxtT1SegDetails `json:"nsxt_t1_seg_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NwSubnetClashDetails *NetworkSubnetClash `json:"nw_subnet_clash_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NwSummarizedDetails *SummarizedInfo `json:"nw_summarized_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OciInfo *OCISetup `json:"oci_info,omitempty"`

	//  Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsAPIVerCheckFailure *OpenStackAPIVersionCheckFailure `json:"os_api_ver_check_failure,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsInfraDetails *OpenStackClusterSetup `json:"os_infra_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsIPDetails *OpenStackIPChange `json:"os_ip_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsLbauditDetails *OpenStackLbProvAuditCheck `json:"os_lbaudit_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsLbpluginOpDetails *OpenStackLbPluginOp `json:"os_lbplugin_op_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsSeVMDetails *OpenStackSeVMChange `json:"os_se_vm_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsSyncServicesDetails *OpenStackSyncServices `json:"os_sync_services_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OsVnicDetails *OpenStackVnicChange `json:"os_vnic_details,omitempty"`

	// PKIProfile event. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PkiprofileDetails *PKIprofileDetails `json:"pkiprofile_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolDeploymentFailureInfo *PoolDeploymentFailureInfo `json:"pool_deployment_failure_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolDeploymentSuccessInfo *PoolDeploymentSuccessInfo `json:"pool_deployment_success_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolDeploymentUpdateInfo *PoolDeploymentUpdateInfo `json:"pool_deployment_update_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PoolServerDeleteDetails *VinfraPoolServerDeleteDetails `json:"pool_server_delete_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PsmProgramDetails *PsmProgramDetails `json:"psm_program_details,omitempty"`

	//  Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RateLimiterEventDetails *RateLimiterEventDetails `json:"rate_limiter_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RebalanceMigrateDetails *RebalanceMigrateEventDetails `json:"rebalance_migrate_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RebalanceScaleinDetails *RebalanceScaleinEventDetails `json:"rebalance_scalein_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RebalanceScaleoutDetails *RebalanceScaleoutEventDetails `json:"rebalance_scaleout_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RebootSeDetails *RmRebootSeEventDetails `json:"reboot_se_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SamlMetadataFailedEvents *SamlMetadataUpdateFailedDetails `json:"saml_metadata_failed_events,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SchedulerActionInfo *SchedulerActionDetails `json:"scheduler_action_info,omitempty"`

	//  Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeBgpPeerDownDetails *SeBgpPeerDownDetails `json:"se_bgp_peer_down_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeBgpPeerStateChangeDetails *SeBgpPeerStateChangeDetails `json:"se_bgp_peer_state_change_details,omitempty"`

	//  Field introduced in 22.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDebugModeEventDetail *SeDebugModeEventDetail `json:"se_debug_mode_event_detail,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDetails *SeMgrEventDetails `json:"se_details,omitempty"`

	//  Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeDiscontinuousTimeChangeEventDetails *SeDiscontinuousTimeChangeEventDetails `json:"se_discontinuous_time_change_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeDupipEventDetails *SeDupipEventDetails `json:"se_dupip_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGatewayHeartbeatFailedDetails *SeGatewayHeartbeatFailedDetails `json:"se_gateway_heartbeat_failed_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGatewayHeartbeatSuccessDetails *SeGatewayHeartbeatSuccessDetails `json:"se_gateway_heartbeat_success_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGeoDbDetails *SeGeoDbDetails `json:"se_geo_db_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHbEventDetails *SeHBEventDetails `json:"se_hb_event_details,omitempty"`

	// Inter-SE datapath heartbeat recovered. One event is generated when heartbeat recovers. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHbRecoveredEventDetails *SeHbRecoveredEventDetails `json:"se_hb_recovered_event_details,omitempty"`

	// Egress queueing latency from proxy to dispatcher. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeHighEgressProcLatencyEventDetails *SeHighEgressProcLatencyEventDetails `json:"se_high_egress_proc_latency_event_details,omitempty"`

	//  Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeHighIngressProcLatencyEventDetails *SeHighIngressProcLatencyEventDetails `json:"se_high_ingress_proc_latency_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHmGsDetails *SeHmEventGSDetails `json:"se_hm_gs_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHmGsgroupDetails *SeHmEventGslbPoolDetails `json:"se_hm_gsgroup_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHmPoolDetails *SeHmEventPoolDetails `json:"se_hm_pool_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeHmVsDetails *SeHmEventVsDetails `json:"se_hm_vs_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeIp6DadFailedEventDetails *SeIp6DadFailedEventDetails `json:"se_ip6_dad_failed_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeIPAddedEventDetails *SeIPAddedEventDetails `json:"se_ip_added_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeIPRemovedEventDetails *SeIPRemovedEventDetails `json:"se_ip_removed_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeIpfailureEventDetails *SeIpfailureEventDetails `json:"se_ipfailure_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeLicensedBandwdithExceededEventDetails *SeLicensedBandwdithExceededEventDetails `json:"se_licensed_bandwdith_exceeded_event_details,omitempty"`

	//  Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMemoryLimitEventDetails *SeMemoryLimitEventDetails `json:"se_memory_limit_event_details,omitempty"`

	// SE NTP synchronization failed. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeNtpSynchronizationFailed *SeNtpSynchronizationFailed `json:"se_ntp_synchronization_failed,omitempty"`

	//  Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeObjsyncPeerDownDetails *SeObjsyncPeerDownDetails `json:"se_objsync_peer_down_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePersistenceDetails *SePersistenceEventDetails `json:"se_persistence_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SePoolLbDetails *SePoolLbEventDetails `json:"se_pool_lb_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeReconcileDetails *SeReconcileDetails `json:"se_reconcile_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeThreshEventDetails *SeThreshEventDetails `json:"se_thresh_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVnicDownEventDetails *SeVnicDownEventDetails `json:"se_vnic_down_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVnicTxQueueStallEventDetails *SeVnicTxQueueStallEventDetails `json:"se_vnic_tx_queue_stall_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVnicUpEventDetails *SeVnicUpEventDetails `json:"se_vnic_up_event_details,omitempty"`

	// VS Flows disrupted when a VS was deleted from SE. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeVsDelFlowsDisrupted *SeVsDelFlowsDisrupted `json:"se_vs_del_flows_disrupted,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeVsFaultEventDetails *SeVsFaultEventDetails `json:"se_vs_fault_event_details,omitempty"`

	//  Field introduced in 18.2.11,20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeVsPktBufHighEventDetails *SeVsPktBufHighEventDetails `json:"se_vs_pkt_buf_high_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecMgrDataEvent *SecMgrDataEvent `json:"sec_mgr_data_event,omitempty"`

	// Security-mgr UA Cache event details. Field introduced in 21.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SecMgrUaEventDetails *SecMgrUAEventDetails `json:"sec_mgr_ua_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecureKeyExchangeInfo *SecureKeyExchangeDetails `json:"secure_key_exchange_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SemigrateEventDetails *SeMigrateEventDetails `json:"semigrate_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerAutoscaleFailedInfo *ServerAutoScaleFailedInfo `json:"server_autoscale_failed_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerAutoscaleinCompleteInfo *ServerAutoScaleInCompleteInfo `json:"server_autoscalein_complete_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerAutoscaleinInfo *ServerAutoScaleInInfo `json:"server_autoscalein_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerAutoscaleoutCompleteInfo *ServerAutoScaleOutCompleteInfo `json:"server_autoscaleout_complete_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ServerAutoscaleoutInfo *ServerAutoScaleOutInfo `json:"server_autoscaleout_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeupgradeDisruptedDetails *SeUpgradeVsDisruptedEventDetails `json:"seupgrade_disrupted_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeupgradeEventDetails *SeUpgradeEventDetails `json:"seupgrade_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeupgradeMigrateDetails *SeUpgradeMigrateEventDetails `json:"seupgrade_migrate_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeupgradeScaleinDetails *SeUpgradeScaleinEventDetails `json:"seupgrade_scalein_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeupgradeScaleoutDetails *SeUpgradeScaleoutEventDetails `json:"seupgrade_scaleout_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SpawnSeDetails *RmSpawnSeEventDetails `json:"spawn_se_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslExpireDetails *SSLExpireDetails `json:"ssl_expire_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslExportDetails *SSLExportDetails `json:"ssl_export_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslIgnoredDetails *SSLIgnoredDetails `json:"ssl_ignored_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslRenewDetails *SSLRenewDetails `json:"ssl_renew_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslRenewFailedDetails *SSLRenewFailedDetails `json:"ssl_renew_failed_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SslRevokedDetails *SSLRevokedDetails `json:"ssl_revoked_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SwitchoverDetails *SwitchoverEventDetails `json:"switchover_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SwitchoverFailDetails *SwitchoverFailEventDetails `json:"switchover_fail_details,omitempty"`

	// Azure cloud sync services event details. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SyncServicesInfo *CloudSyncServices `json:"sync_services_info,omitempty"`

	// System Report event details. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SystemReportEventDetails *SystemReport `json:"system_report_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TencentInfo *TencentSetup `json:"tencent_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UnbindVsSeDetails *RmUnbindVsSeEventDetails `json:"unbind_vs_se_details,omitempty"`

	//  Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpgradeEntry *UpgradeOpsEntry `json:"upgrade_entry,omitempty"`

	//  Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpgradeStatusInfo *UpgradeStatusInfo `json:"upgrade_status_info,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcaInfraDetails *VCASetup `json:"vca_infra_details,omitempty"`

	// Details of objects still referred to cloud. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterCloudDeleteDetails *VcenterCloudDeleteDetails `json:"vcenter_cloud_delete_details,omitempty"`

	// VCenter Cluster event. Field introduced in 20.1.7, 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterClusterDetails *VcenterClusterDetails `json:"vcenter_cluster_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterConnectivityStatus *VinfraVcenterConnectivityStatus `json:"vcenter_connectivity_status,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterDetails *VinfraVcenterBadCredentials `json:"vcenter_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterDiscFailure *VinfraVcenterDiscoveryFailure `json:"vcenter_disc_failure,omitempty"`

	// Vcenter Image event details. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterImgDetails *VcenterImageDetails `json:"vcenter_img_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterNetworkLimit *VinfraVcenterNetworkLimit `json:"vcenter_network_limit,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterObjDeleteDetails *VinfraVcenterObjDeleteDetails `json:"vcenter_obj_delete_details,omitempty"`

	// Failed to tag SEs with custom tags. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterTagEventDetails *VcenterTagEventDetails `json:"vcenter_tag_event_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipAutoscale *VipScaleDetails `json:"vip_autoscale,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VipDNSInfo *DNSRegisterInfo `json:"vip_dns_info,omitempty"`

	// Details for VIP Symmetry. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VipSymmetryDetails *VipSymmetryDetails `json:"vip_symmetry_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VMDetails *VinfraVMDetails `json:"vm_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsAwaitingseDetails *VsAwaitingSeEventDetails `json:"vs_awaitingse_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsErrorDetails *VsErrorEventDetails `json:"vs_error_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsFsmDetails *VsFsmEventDetails `json:"vs_fsm_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsInitialplacementDetails *VsInitialPlacementEventDetails `json:"vs_initialplacement_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsMigrateDetails *VsMigrateEventDetails `json:"vs_migrate_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsPoolNwFltrDetails *VsPoolNwFilterEventDetails `json:"vs_pool_nw_fltr_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsScaleinDetails *VsScaleInEventDetails `json:"vs_scalein_details,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VsScaleoutDetails *VsScaleOutEventDetails `json:"vs_scaleout_details,omitempty"`

	// Details for Primary Switchover status. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsSwitchoverDetails *VsSwitchoverEventDetails `json:"vs_switchover_details,omitempty"`
}
