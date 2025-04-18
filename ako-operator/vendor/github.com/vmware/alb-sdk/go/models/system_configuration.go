// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SystemConfiguration system configuration
// swagger:model SystemConfiguration
type SystemConfiguration struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AdminAuthConfiguration *AdminAuthConfiguration `json:"admin_auth_configuration,omitempty"`

	// Common criteria mode's current state. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CommonCriteriaMode *bool `json:"common_criteria_mode,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Controller metrics event dynamic thresholds can be set here. CONTROLLER_CPU_HIGH and CONTROLLER_MEM_HIGH evets can take configured dynamic thresholds. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ControllerAnalyticsPolicy *ControllerAnalyticsPolicy `json:"controller_analytics_policy,omitempty"`

	// Specifies the default license tier which would be used by new Clouds. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS, ENTERPRISE_WITH_CLOUD_SERVICES. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition. Special default for Essentials edition is ESSENTIALS, Basic edition is BASIC, Enterprise is ENTERPRISE_WITH_CLOUD_SERVICES.
	DefaultLicenseTier *string `json:"default_license_tier,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DNSConfiguration *DNSConfiguration `json:"dns_configuration,omitempty"`

	// DNS virtualservices hosting FQDN records for applications across Avi Vantage. If no virtualservices are provided, Avi Vantage will provide DNS services for configured applications. Switching back to Avi Vantage from DNS virtualservices is not allowed. It is a reference to an object of type VirtualService. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSVirtualserviceRefs []string `json:"dns_virtualservice_refs,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DockerMode *bool `json:"docker_mode,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EmailConfiguration *EmailConfiguration `json:"email_configuration,omitempty"`

	// Enable CORS Header. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	EnableCors *bool `json:"enable_cors,omitempty"`

	// FIPS mode current state. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FipsMode *bool `json:"fips_mode,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GlobalTenantConfig *TenantConfiguration `json:"global_tenant_config,omitempty"`

	// Users can specify comma separated list of deprecated host key algorithm.If nothing is specified, all known algorithms provided by OpenSSH will be supported.This change could only apply on the controller node. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostKeyAlgorithmExclude *string `json:"host_key_algorithm_exclude,omitempty"`

	// Users can specify comma separated list of deprecated key exchange algorithm.If nothing is specified, all known algorithms provided by OpenSSH will be supported.This change could only apply on the controller node. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KexAlgorithmExclude *string `json:"kex_algorithm_exclude,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LinuxConfiguration *LinuxConfiguration `json:"linux_configuration,omitempty"`

	// Configure Ip Access control for controller to restrict open access. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MgmtIPAccessControl *MgmtIPAccessControl `json:"mgmt_ip_access_control,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NtpConfiguration *NTPConfiguration `json:"ntp_configuration,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PortalConfiguration *PortalConfiguration `json:"portal_configuration,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ProxyConfiguration *ProxyConfiguration `json:"proxy_configuration,omitempty"`

	// Users can specify and update the time limit of RekeyLimit in sshd_config.If nothing is specified, the default setting will be none. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RekeyTimeLimit *string `json:"rekey_time_limit,omitempty"`

	// Users can specify and update the size/volume limit of RekeyLimit in sshd_config.If nothing is specified, the default setting will be default. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RekeyVolumeLimit *string `json:"rekey_volume_limit,omitempty"`

	// Configure Secure Channel properties. Field introduced in 18.1.4, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SecureChannelConfiguration *SecureChannelConfiguration `json:"secure_channel_configuration,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SnmpConfiguration *SnmpConfiguration `json:"snmp_configuration,omitempty"`

	// Allowed Ciphers list for SSH to the management interface on the Controller and Service Engines. If this is not specified, all the default ciphers are allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SSHCiphers []string `json:"ssh_ciphers,omitempty"`

	// Allowed HMAC list for SSH to the management interface on the Controller and Service Engines. If this is not specified, all the default HMACs are allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SSHHmacs []string `json:"ssh_hmacs,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// This flag is set once the Initial Controller Setup workflow is complete. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	WelcomeWorkflowComplete *bool `json:"welcome_workflow_complete,omitempty"`
}
