package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SystemConfiguration system configuration
// swagger:model SystemConfiguration
type SystemConfiguration struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Placeholder for description of property admin_auth_configuration of obj type SystemConfiguration field type str  type object
	AdminAuthConfiguration *AdminAuthConfiguration `json:"admin_auth_configuration,omitempty"`

	// Enable Common Criteria Compliance mode (Disabled by default). WARN  Toggling this field is disruptive and will result in reduced behavior with SSH and TLS protocols. Expect possible warm start of control and data planes. Field introduced in 20.1.3.
	CommonCriteriaMode *bool `json:"common_criteria_mode,omitempty"`

	// Specifies the default license tier which would be used by new Clouds. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS. Field introduced in 17.2.5. Allowed in Basic edition, Essentials edition, Enterprise edition. Special default for Basic edition is BASIC, Essentials edition is ESSENTIALS, Enterprise is ENTERPRISE.
	DefaultLicenseTier *string `json:"default_license_tier,omitempty"`

	// Placeholder for description of property dns_configuration of obj type SystemConfiguration field type str  type object
	DNSConfiguration *DNSConfiguration `json:"dns_configuration,omitempty"`

	// DNS virtualservices hosting FQDN records for applications across Avi Vantage. If no virtualservices are provided, Avi Vantage will provide DNS services for configured applications. Switching back to Avi Vantage from DNS virtualservices is not allowed. It is a reference to an object of type VirtualService.
	DNSVirtualserviceRefs []string `json:"dns_virtualservice_refs,omitempty"`

	// Placeholder for description of property docker_mode of obj type SystemConfiguration field type str  type boolean
	DockerMode *bool `json:"docker_mode,omitempty"`

	// Placeholder for description of property email_configuration of obj type SystemConfiguration field type str  type object
	EmailConfiguration *EmailConfiguration `json:"email_configuration,omitempty"`

	// Enable CORS Header. Field introduced in 20.1.3. Allowed in Basic edition, Essentials edition, Enterprise edition.
	EnableCors *bool `json:"enable_cors,omitempty"`

	// Enable FIPS mode. Field introduced in 20.1.1.
	FipsMode *bool `json:"fips_mode,omitempty"`

	// Placeholder for description of property global_tenant_config of obj type SystemConfiguration field type str  type object
	GlobalTenantConfig *TenantConfiguration `json:"global_tenant_config,omitempty"`

	// Placeholder for description of property linux_configuration of obj type SystemConfiguration field type str  type object
	LinuxConfiguration *LinuxConfiguration `json:"linux_configuration,omitempty"`

	// Configure Ip Access control for controller to restrict open access.
	MgmtIPAccessControl *MgmtIPAccessControl `json:"mgmt_ip_access_control,omitempty"`

	// Placeholder for description of property ntp_configuration of obj type SystemConfiguration field type str  type object
	NtpConfiguration *NTPConfiguration `json:"ntp_configuration,omitempty"`

	// Placeholder for description of property portal_configuration of obj type SystemConfiguration field type str  type object
	PortalConfiguration *PortalConfiguration `json:"portal_configuration,omitempty"`

	// Placeholder for description of property proxy_configuration of obj type SystemConfiguration field type str  type object
	ProxyConfiguration *ProxyConfiguration `json:"proxy_configuration,omitempty"`

	// Configure Secure Channel properties. Field introduced in 18.1.4, 18.2.1.
	SecureChannelConfiguration *SecureChannelConfiguration `json:"secure_channel_configuration,omitempty"`

	// Placeholder for description of property snmp_configuration of obj type SystemConfiguration field type str  type object
	SnmpConfiguration *SnmpConfiguration `json:"snmp_configuration,omitempty"`

	// Allowed Ciphers list for SSH to the management interface on the Controller and Service Engines. If this is not specified, all the default ciphers are allowed.
	SSHCiphers []string `json:"ssh_ciphers,omitempty"`

	// Allowed HMAC list for SSH to the management interface on the Controller and Service Engines. If this is not specified, all the default HMACs are allowed.
	SSHHmacs []string `json:"ssh_hmacs,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// This flag is set once the Initial Controller Setup workflow is complete. Field introduced in 18.2.3.
	WelcomeWorkflowComplete *bool `json:"welcome_workflow_complete,omitempty"`
}
