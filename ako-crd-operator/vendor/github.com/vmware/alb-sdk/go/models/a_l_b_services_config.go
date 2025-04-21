// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ALBServicesConfig a l b services config
// swagger:model ALBServicesConfig
type ALBServicesConfig struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Default values for Application Signature sync. Field introduced in 20.1.4. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	AppSignatureConfig *AppSignatureConfig `json:"app_signature_config"`

	// Default contact for this controller cluster. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AssetContact *ALBServicesUser `json:"asset_contact,omitempty"`

	// Default values for case management. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	CaseConfig *CaseConfig `json:"case_config"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Features opt-in for Pulse Cloud Services. Field introduced in 20.1.1.
	// Required: true
	FeatureOptInStatus *PortalFeatureOptIn `json:"feature_opt_in_status"`

	// Default values to be used for IP Reputation sync. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	IPReputationConfig *IPReputationConfig `json:"ip_reputation_config"`

	// Mode helps log collection and upload. Enum options - MODE_UNKNOWN, SALESFORCE, SYSTEST, MYVMWARE. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- SALESFORCE,MYVMWARE,SYSTEST), Basic edition(Allowed values- SALESFORCE,MYVMWARE,SYSTEST), Enterprise with Cloud Services edition.
	Mode *string `json:"mode,omitempty"`

	// Name of the ALBServicesConfig object. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Time interval in minutes. Allowed values are 5-60. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PollingInterval *uint32 `json:"polling_interval,omitempty"`

	// The FQDN or IP address of the Pulse Cloud Services. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	PortalURL *string `json:"portal_url"`

	// Saas licensing configuration. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	SaasLicensingConfig *SaasLicensingInfo `json:"saas_licensing_config"`

	// Session configuration data. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	SessionConfig *PulseServicesSessionConfig `json:"session_config,omitempty"`

	// Split proxy configuration to connect external Pulse Cloud Services. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	SplitProxyConfiguration *ProxyConfiguration `json:"split_proxy_configuration"`

	// Tenant based configuration data. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	TenantConfig *PulseServicesTenantConfig `json:"tenant_config,omitempty"`

	// Tenant UUID associated with the Object. It is a reference to an object of type Tenant. Field introduced in 30.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// By default, Pulse Cloud Services uses proxy added in system configuration. If it should use a separate proxy, set this flag to true and configure split proxy configuration. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseSplitProxy *bool `json:"use_split_proxy,omitempty"`

	// Secure the controller to Pulse Cloud Services communication over TLS. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	UseTLS *bool `json:"use_tls,omitempty"`

	// Default values for user agent DB Service. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	UserAgentDbConfig *UserAgentDBConfig `json:"user_agent_db_config"`

	//  Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// Default values for waf management. Field introduced in 21.1.1. Allowed in Essentials edition with any value, Basic edition with any value, Enterprise, Enterprise with Cloud Services edition.
	// Required: true
	WafConfig *WafCrsConfig `json:"waf_config"`
}
