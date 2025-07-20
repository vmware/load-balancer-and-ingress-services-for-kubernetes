// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0

package clients

import (
	"github.com/vmware/alb-sdk/go/session"
)

// AviClient -- an API Client for Avi Controller
type AviClient struct {
	AviSession                      *session.AviSession
	ALBServicesConfig               *ALBServicesConfigClient
	ALBServicesFileDownload         *ALBServicesFileDownloadClient
	ALBServicesFileUpload           *ALBServicesFileUploadClient
	ALBServicesJob                  *ALBServicesJobClient
	ActionGroupConfig               *ActionGroupConfigClient
	Alert                           *AlertClient
	AlertConfig                     *AlertConfigClient
	AlertEmailConfig                *AlertEmailConfigClient
	AlertObjectList                 *AlertObjectListClient
	AlertScriptConfig               *AlertScriptConfigClient
	AlertSyslogConfig               *AlertSyslogConfigClient
	AnalyticsProfile                *AnalyticsProfileClient
	Application                     *ApplicationClient
	ApplicationPersistenceProfile   *ApplicationPersistenceProfileClient
	ApplicationProfile              *ApplicationProfileClient
	AuthMappingProfile              *AuthMappingProfileClient
	AuthProfile                     *AuthProfileClient
	AutoScaleLaunchConfig           *AutoScaleLaunchConfigClient
	AvailabilityZone                *AvailabilityZoneClient
	Backup                          *BackupClient
	BackupConfiguration             *BackupConfigurationClient
	BotConfigConsolidator           *BotConfigConsolidatorClient
	BotDetectionPolicy              *BotDetectionPolicyClient
	BotIPReputationTypeMapping      *BotIPReputationTypeMappingClient
	BotMapping                      *BotMappingClient
	CSRFPolicy                      *CSRFPolicyClient
	CertificateManagementProfile    *CertificateManagementProfileClient
	Cloud                           *CloudClient
	CloudConnectorUser              *CloudConnectorUserClient
	CloudInventory                  *CloudInventoryClient
	CloudProperties                 *CloudPropertiesClient
	CloudRuntime                    *CloudRuntimeClient
	ClusterCloudDetails             *ClusterCloudDetailsClient
	ControllerPortalRegistration    *ControllerPortalRegistrationClient
	ControllerProperties            *ControllerPropertiesClient
	ControllerSite                  *ControllerSiteClient
	CustomIPAMDNSProfile            *CustomIPAMDNSProfileClient
	DNSPolicy                       *DNSPolicyClient
	DebugController                 *DebugControllerClient
	DebugServiceEngine              *DebugServiceEngineClient
	DebugVirtualService             *DebugVirtualServiceClient
	DynamicDNSRecord                *DynamicDNSRecordClient
	ErrorPageBody                   *ErrorPageBodyClient
	ErrorPageProfile                *ErrorPageProfileClient
	FederationCheckpoint            *FederationCheckpointClient
	FederationCheckpointInventory   *FederationCheckpointInventoryClient
	FileObject                      *FileObjectClient
	Generic                         *GenericClient
	GeoDB                           *GeoDBClient
	Gslb                            *GslbClient
	GslbGeoDbProfile                *GslbGeoDbProfileClient
	GslbInventory                   *GslbInventoryClient
	GslbService                     *GslbServiceClient
	GslbServiceInventory            *GslbServiceInventoryClient
	HTTPPolicySet                   *HTTPPolicySetClient
	HardwareSecurityModuleGroup     *HardwareSecurityModuleGroupClient
	HealthMonitor                   *HealthMonitorClient
	IPAMDNSProviderProfile          *IPAMDNSProviderProfileClient
	IPAddrGroup                     *IPAddrGroupClient
	IPReputationDB                  *IPReputationDBClient
	IcapProfile                     *IcapProfileClient
	Image                           *ImageClient
	InventoryFaultConfig            *InventoryFaultConfigClient
	JWTServerProfile                *JWTServerProfileClient
	JobEntry                        *JobEntryClient
	L4PolicySet                     *L4PolicySetClient
	LabelGroup                      *LabelGroupClient
	LicenseLedgerDetails            *LicenseLedgerDetailsClient
	LicenseStatus                   *LicenseStatusClient
	LogControllerMapping            *LogControllerMappingClient
	MemoryBalancerRequest           *MemoryBalancerRequestClient
	MicroService                    *MicroServiceClient
	MicroServiceGroup               *MicroServiceGroupClient
	NatPolicy                       *NatPolicyClient
	Network                         *NetworkClient
	NetworkInventory                *NetworkInventoryClient
	NetworkProfile                  *NetworkProfileClient
	NetworkRuntime                  *NetworkRuntimeClient
	NetworkSecurityPolicy           *NetworkSecurityPolicyClient
	NetworkService                  *NetworkServiceClient
	NsxtSegmentRuntime              *NsxtSegmentRuntimeClient
	PKIprofile                      *PKIprofileClient
	Pool                            *PoolClient
	PoolGroup                       *PoolGroupClient
	PoolGroupDeploymentPolicy       *PoolGroupDeploymentPolicyClient
	PoolGroupInventory              *PoolGroupInventoryClient
	PoolInventory                   *PoolInventoryClient
	PriorityLabels                  *PriorityLabelsClient
	ProtocolParser                  *ProtocolParserClient
	Role                            *RoleClient
	SCPoolServerStateInfo           *SCPoolServerStateInfoClient
	SCVsStateInfo                   *SCVsStateInfoClient
	SSLKeyAndCertificate            *SSLKeyAndCertificateClient
	SSLProfile                      *SSLProfileClient
	SSOPolicy                       *SSOPolicyClient
	Scheduler                       *SchedulerClient
	SeProperties                    *SePropertiesClient
	SecureChannelAvailableLocalIps  *SecureChannelAvailableLocalIpsClient
	SecureChannelMapping            *SecureChannelMappingClient
	SecureChannelToken              *SecureChannelTokenClient
	SecurityManagerData             *SecurityManagerDataClient
	SecurityPolicy                  *SecurityPolicyClient
	ServerAutoScalePolicy           *ServerAutoScalePolicyClient
	ServiceAuthProfile              *ServiceAuthProfileClient
	ServiceEngine                   *ServiceEngineClient
	ServiceEngineGroup              *ServiceEngineGroupClient
	ServiceEngineGroupInventory     *ServiceEngineGroupInventoryClient
	ServiceEngineInventory          *ServiceEngineInventoryClient
	SiteVersion                     *SiteVersionClient
	SnmpTrapProfile                 *SnmpTrapProfileClient
	StatediffOperation              *StatediffOperationClient
	StatediffSnapshot               *StatediffSnapshotClient
	StringGroup                     *StringGroupClient
	SystemConfiguration             *SystemConfigurationClient
	SystemLimits                    *SystemLimitsClient
	SystemReport                    *SystemReportClient
	TaskJournal                     *TaskJournalClient
	Tenant                          *TenantClient
	TenantSystemConfiguration       *TenantSystemConfigurationClient
	TestSeDatastoreLevel1           *TestSeDatastoreLevel1Client
	TestSeDatastoreLevel2           *TestSeDatastoreLevel2Client
	TestSeDatastoreLevel3           *TestSeDatastoreLevel3Client
	TrafficCloneProfile             *TrafficCloneProfileClient
	UpgradeStatusInfo               *UpgradeStatusInfoClient
	UpgradeStatusSummary            *UpgradeStatusSummaryClient
	User                            *UserClient
	UserAccountProfile              *UserAccountProfileClient
	UserActivity                    *UserActivityClient
	VCenterServer                   *VCenterServerClient
	VIDCInfo                        *VIDCInfoClient
	VIMgrClusterRuntime             *VIMgrClusterRuntimeClient
	VIMgrHostRuntime                *VIMgrHostRuntimeClient
	VIMgrNWRuntime                  *VIMgrNWRuntimeClient
	VIMgrSEVMRuntime                *VIMgrSEVMRuntimeClient
	VIMgrVMRuntime                  *VIMgrVMRuntimeClient
	VIPGNameInfo                    *VIPGNameInfoClient
	VSDataScriptSet                 *VSDataScriptSetClient
	VirtualService                  *VirtualServiceClient
	VrfContext                      *VrfContextClient
	VsGs                            *VsGsClient
	VsInventory                     *VsInventoryClient
	VsVip                           *VsVipClient
	VsvipInventory                  *VsvipInventoryClient
	WafApplicationSignatureProvider *WafApplicationSignatureProviderClient
	WafCRS                          *WafCRSClient
	WafPolicy                       *WafPolicyClient
	WafPolicyPSMGroup               *WafPolicyPSMGroupClient
	WafPolicyPSMGroupInventory      *WafPolicyPSMGroupInventoryClient
	WafProfile                      *WafProfileClient
	WebappUT                        *WebappUTClient
	Webhook                         *WebhookClient
}

// NewAviClient initiates an AviSession and returns an AviClient wrapping that session
func NewAviClient(host string, username string, options ...func(*session.AviSession) error) (*AviClient, error) {
	aviClient := AviClient{}
	aviSession, err := session.NewAviSession(host, username, options...)
	if err != nil {
		return &aviClient, err
	}
	aviClient.AviSession = aviSession
	aviClient.ALBServicesConfig = NewALBServicesConfigClient(aviSession)
	aviClient.ALBServicesFileDownload = NewALBServicesFileDownloadClient(aviSession)
	aviClient.ALBServicesFileUpload = NewALBServicesFileUploadClient(aviSession)
	aviClient.ALBServicesJob = NewALBServicesJobClient(aviSession)
	aviClient.ActionGroupConfig = NewActionGroupConfigClient(aviSession)
	aviClient.Alert = NewAlertClient(aviSession)
	aviClient.AlertConfig = NewAlertConfigClient(aviSession)
	aviClient.AlertEmailConfig = NewAlertEmailConfigClient(aviSession)
	aviClient.AlertObjectList = NewAlertObjectListClient(aviSession)
	aviClient.AlertScriptConfig = NewAlertScriptConfigClient(aviSession)
	aviClient.AlertSyslogConfig = NewAlertSyslogConfigClient(aviSession)
	aviClient.AnalyticsProfile = NewAnalyticsProfileClient(aviSession)
	aviClient.Application = NewApplicationClient(aviSession)
	aviClient.ApplicationPersistenceProfile = NewApplicationPersistenceProfileClient(aviSession)
	aviClient.ApplicationProfile = NewApplicationProfileClient(aviSession)
	aviClient.AuthMappingProfile = NewAuthMappingProfileClient(aviSession)
	aviClient.AuthProfile = NewAuthProfileClient(aviSession)
	aviClient.AutoScaleLaunchConfig = NewAutoScaleLaunchConfigClient(aviSession)
	aviClient.AvailabilityZone = NewAvailabilityZoneClient(aviSession)
	aviClient.Backup = NewBackupClient(aviSession)
	aviClient.BackupConfiguration = NewBackupConfigurationClient(aviSession)
	aviClient.BotConfigConsolidator = NewBotConfigConsolidatorClient(aviSession)
	aviClient.BotDetectionPolicy = NewBotDetectionPolicyClient(aviSession)
	aviClient.BotIPReputationTypeMapping = NewBotIPReputationTypeMappingClient(aviSession)
	aviClient.BotMapping = NewBotMappingClient(aviSession)
	aviClient.CSRFPolicy = NewCSRFPolicyClient(aviSession)
	aviClient.CertificateManagementProfile = NewCertificateManagementProfileClient(aviSession)
	aviClient.Cloud = NewCloudClient(aviSession)
	aviClient.CloudConnectorUser = NewCloudConnectorUserClient(aviSession)
	aviClient.CloudInventory = NewCloudInventoryClient(aviSession)
	aviClient.CloudProperties = NewCloudPropertiesClient(aviSession)
	aviClient.CloudRuntime = NewCloudRuntimeClient(aviSession)
	aviClient.ClusterCloudDetails = NewClusterCloudDetailsClient(aviSession)
	aviClient.ControllerPortalRegistration = NewControllerPortalRegistrationClient(aviSession)
	aviClient.ControllerProperties = NewControllerPropertiesClient(aviSession)
	aviClient.ControllerSite = NewControllerSiteClient(aviSession)
	aviClient.CustomIPAMDNSProfile = NewCustomIPAMDNSProfileClient(aviSession)
	aviClient.DNSPolicy = NewDNSPolicyClient(aviSession)
	aviClient.DebugController = NewDebugControllerClient(aviSession)
	aviClient.DebugServiceEngine = NewDebugServiceEngineClient(aviSession)
	aviClient.DebugVirtualService = NewDebugVirtualServiceClient(aviSession)
	aviClient.DynamicDNSRecord = NewDynamicDNSRecordClient(aviSession)
	aviClient.ErrorPageBody = NewErrorPageBodyClient(aviSession)
	aviClient.ErrorPageProfile = NewErrorPageProfileClient(aviSession)
	aviClient.FederationCheckpoint = NewFederationCheckpointClient(aviSession)
	aviClient.FederationCheckpointInventory = NewFederationCheckpointInventoryClient(aviSession)
	aviClient.FileObject = NewFileObjectClient(aviSession)
	aviClient.Generic = NewGenericClient(aviSession)
	aviClient.GeoDB = NewGeoDBClient(aviSession)
	aviClient.Gslb = NewGslbClient(aviSession)
	aviClient.GslbGeoDbProfile = NewGslbGeoDbProfileClient(aviSession)
	aviClient.GslbInventory = NewGslbInventoryClient(aviSession)
	aviClient.GslbService = NewGslbServiceClient(aviSession)
	aviClient.GslbServiceInventory = NewGslbServiceInventoryClient(aviSession)
	aviClient.HTTPPolicySet = NewHTTPPolicySetClient(aviSession)
	aviClient.HardwareSecurityModuleGroup = NewHardwareSecurityModuleGroupClient(aviSession)
	aviClient.HealthMonitor = NewHealthMonitorClient(aviSession)
	aviClient.IPAMDNSProviderProfile = NewIPAMDNSProviderProfileClient(aviSession)
	aviClient.IPAddrGroup = NewIPAddrGroupClient(aviSession)
	aviClient.IPReputationDB = NewIPReputationDBClient(aviSession)
	aviClient.IcapProfile = NewIcapProfileClient(aviSession)
	aviClient.Image = NewImageClient(aviSession)
	aviClient.InventoryFaultConfig = NewInventoryFaultConfigClient(aviSession)
	aviClient.JWTServerProfile = NewJWTServerProfileClient(aviSession)
	aviClient.JobEntry = NewJobEntryClient(aviSession)
	aviClient.L4PolicySet = NewL4PolicySetClient(aviSession)
	aviClient.LabelGroup = NewLabelGroupClient(aviSession)
	aviClient.LicenseLedgerDetails = NewLicenseLedgerDetailsClient(aviSession)
	aviClient.LicenseStatus = NewLicenseStatusClient(aviSession)
	aviClient.LogControllerMapping = NewLogControllerMappingClient(aviSession)
	aviClient.MemoryBalancerRequest = NewMemoryBalancerRequestClient(aviSession)
	aviClient.MicroService = NewMicroServiceClient(aviSession)
	aviClient.MicroServiceGroup = NewMicroServiceGroupClient(aviSession)
	aviClient.NatPolicy = NewNatPolicyClient(aviSession)
	aviClient.Network = NewNetworkClient(aviSession)
	aviClient.NetworkInventory = NewNetworkInventoryClient(aviSession)
	aviClient.NetworkProfile = NewNetworkProfileClient(aviSession)
	aviClient.NetworkRuntime = NewNetworkRuntimeClient(aviSession)
	aviClient.NetworkSecurityPolicy = NewNetworkSecurityPolicyClient(aviSession)
	aviClient.NetworkService = NewNetworkServiceClient(aviSession)
	aviClient.NsxtSegmentRuntime = NewNsxtSegmentRuntimeClient(aviSession)
	aviClient.PKIprofile = NewPKIprofileClient(aviSession)
	aviClient.Pool = NewPoolClient(aviSession)
	aviClient.PoolGroup = NewPoolGroupClient(aviSession)
	aviClient.PoolGroupDeploymentPolicy = NewPoolGroupDeploymentPolicyClient(aviSession)
	aviClient.PoolGroupInventory = NewPoolGroupInventoryClient(aviSession)
	aviClient.PoolInventory = NewPoolInventoryClient(aviSession)
	aviClient.PriorityLabels = NewPriorityLabelsClient(aviSession)
	aviClient.ProtocolParser = NewProtocolParserClient(aviSession)
	aviClient.Role = NewRoleClient(aviSession)
	aviClient.SCPoolServerStateInfo = NewSCPoolServerStateInfoClient(aviSession)
	aviClient.SCVsStateInfo = NewSCVsStateInfoClient(aviSession)
	aviClient.SSLKeyAndCertificate = NewSSLKeyAndCertificateClient(aviSession)
	aviClient.SSLProfile = NewSSLProfileClient(aviSession)
	aviClient.SSOPolicy = NewSSOPolicyClient(aviSession)
	aviClient.Scheduler = NewSchedulerClient(aviSession)
	aviClient.SeProperties = NewSePropertiesClient(aviSession)
	aviClient.SecureChannelAvailableLocalIps = NewSecureChannelAvailableLocalIpsClient(aviSession)
	aviClient.SecureChannelMapping = NewSecureChannelMappingClient(aviSession)
	aviClient.SecureChannelToken = NewSecureChannelTokenClient(aviSession)
	aviClient.SecurityManagerData = NewSecurityManagerDataClient(aviSession)
	aviClient.SecurityPolicy = NewSecurityPolicyClient(aviSession)
	aviClient.ServerAutoScalePolicy = NewServerAutoScalePolicyClient(aviSession)
	aviClient.ServiceAuthProfile = NewServiceAuthProfileClient(aviSession)
	aviClient.ServiceEngine = NewServiceEngineClient(aviSession)
	aviClient.ServiceEngineGroup = NewServiceEngineGroupClient(aviSession)
	aviClient.ServiceEngineGroupInventory = NewServiceEngineGroupInventoryClient(aviSession)
	aviClient.ServiceEngineInventory = NewServiceEngineInventoryClient(aviSession)
	aviClient.SiteVersion = NewSiteVersionClient(aviSession)
	aviClient.SnmpTrapProfile = NewSnmpTrapProfileClient(aviSession)
	aviClient.StatediffOperation = NewStatediffOperationClient(aviSession)
	aviClient.StatediffSnapshot = NewStatediffSnapshotClient(aviSession)
	aviClient.StringGroup = NewStringGroupClient(aviSession)
	aviClient.SystemConfiguration = NewSystemConfigurationClient(aviSession)
	aviClient.SystemLimits = NewSystemLimitsClient(aviSession)
	aviClient.SystemReport = NewSystemReportClient(aviSession)
	aviClient.TaskJournal = NewTaskJournalClient(aviSession)
	aviClient.Tenant = NewTenantClient(aviSession)
	aviClient.TenantSystemConfiguration = NewTenantSystemConfigurationClient(aviSession)
	aviClient.TestSeDatastoreLevel1 = NewTestSeDatastoreLevel1Client(aviSession)
	aviClient.TestSeDatastoreLevel2 = NewTestSeDatastoreLevel2Client(aviSession)
	aviClient.TestSeDatastoreLevel3 = NewTestSeDatastoreLevel3Client(aviSession)
	aviClient.TrafficCloneProfile = NewTrafficCloneProfileClient(aviSession)
	aviClient.UpgradeStatusInfo = NewUpgradeStatusInfoClient(aviSession)
	aviClient.UpgradeStatusSummary = NewUpgradeStatusSummaryClient(aviSession)
	aviClient.User = NewUserClient(aviSession)
	aviClient.UserAccountProfile = NewUserAccountProfileClient(aviSession)
	aviClient.UserActivity = NewUserActivityClient(aviSession)
	aviClient.VCenterServer = NewVCenterServerClient(aviSession)
	aviClient.VIDCInfo = NewVIDCInfoClient(aviSession)
	aviClient.VIMgrClusterRuntime = NewVIMgrClusterRuntimeClient(aviSession)
	aviClient.VIMgrHostRuntime = NewVIMgrHostRuntimeClient(aviSession)
	aviClient.VIMgrNWRuntime = NewVIMgrNWRuntimeClient(aviSession)
	aviClient.VIMgrSEVMRuntime = NewVIMgrSEVMRuntimeClient(aviSession)
	aviClient.VIMgrVMRuntime = NewVIMgrVMRuntimeClient(aviSession)
	aviClient.VIPGNameInfo = NewVIPGNameInfoClient(aviSession)
	aviClient.VSDataScriptSet = NewVSDataScriptSetClient(aviSession)
	aviClient.VirtualService = NewVirtualServiceClient(aviSession)
	aviClient.VrfContext = NewVrfContextClient(aviSession)
	aviClient.VsGs = NewVsGsClient(aviSession)
	aviClient.VsInventory = NewVsInventoryClient(aviSession)
	aviClient.VsVip = NewVsVipClient(aviSession)
	aviClient.VsvipInventory = NewVsvipInventoryClient(aviSession)
	aviClient.WafApplicationSignatureProvider = NewWafApplicationSignatureProviderClient(aviSession)
	aviClient.WafCRS = NewWafCRSClient(aviSession)
	aviClient.WafPolicy = NewWafPolicyClient(aviSession)
	aviClient.WafPolicyPSMGroup = NewWafPolicyPSMGroupClient(aviSession)
	aviClient.WafPolicyPSMGroupInventory = NewWafPolicyPSMGroupInventoryClient(aviSession)
	aviClient.WafProfile = NewWafProfileClient(aviSession)
	aviClient.WebappUT = NewWebappUTClient(aviSession)
	aviClient.Webhook = NewWebhookClient(aviSession)
	return &aviClient, nil
}
