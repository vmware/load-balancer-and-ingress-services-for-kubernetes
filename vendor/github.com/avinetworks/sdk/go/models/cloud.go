package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// Cloud cloud
// swagger:model Cloud
type Cloud struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Placeholder for description of property apic_configuration of obj type Cloud field type str  type object
	ApicConfiguration *APICConfiguration `json:"apic_configuration,omitempty"`

	//  Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	ApicMode *bool `json:"apic_mode,omitempty"`

	// CloudConnector polling interval in seconds for external autoscale groups, minimum 60 seconds. Allowed values are 60-3600. Field introduced in 18.2.2. Unit is SECONDS. Allowed in Basic(Allowed values- 60) edition, Essentials(Allowed values- 60) edition, Enterprise edition.
	AutoscalePollingInterval *int32 `json:"autoscale_polling_interval,omitempty"`

	// Placeholder for description of property aws_configuration of obj type Cloud field type str  type object
	AwsConfiguration *AwsConfiguration `json:"aws_configuration,omitempty"`

	//  Field introduced in 17.2.1. Allowed in Basic edition, Essentials edition, Enterprise edition.
	AzureConfiguration *AzureConfiguration `json:"azure_configuration,omitempty"`

	// Placeholder for description of property cloudstack_configuration of obj type Cloud field type str  type object
	CloudstackConfiguration *CloudStackConfiguration `json:"cloudstack_configuration,omitempty"`

	// Custom tags for all Avi created resources in the cloud infrastructure. Field introduced in 17.1.5.
	CustomTags []*CustomTag `json:"custom_tags,omitempty"`

	// Select the IP address management scheme.
	DhcpEnabled *bool `json:"dhcp_enabled,omitempty"`

	// DNS Profile for the cloud. It is a reference to an object of type IpamDnsProviderProfile.
	DNSProviderRef *string `json:"dns_provider_ref,omitempty"`

	// By default, pool member FQDNs are resolved on the Controller. When this is set, pool member FQDNs are instead resolved on Service Engines in this cloud. This is useful in scenarios where pool member FQDNs can only be resolved from Service Engines and not from the Controller. Field introduced in 18.2.6. Allowed in Basic(Allowed values- false) edition, Essentials(Allowed values- false) edition, Enterprise edition.
	DNSResolutionOnSe *bool `json:"dns_resolution_on_se,omitempty"`

	// Placeholder for description of property docker_configuration of obj type Cloud field type str  type object
	DockerConfiguration *DockerConfiguration `json:"docker_configuration,omitempty"`

	// DNS Profile for East-West services. It is a reference to an object of type IpamDnsProviderProfile.
	EastWestDNSProviderRef *string `json:"east_west_dns_provider_ref,omitempty"`

	// Ipam Profile for East-West services. Warning - Please use virtual subnets in this IPAM profile that do not conflict with the underlay networks or any overlay networks in the cluster. For example in AWS and GCP, 169.254.0.0/16 is used for storing instance metadata. Hence, it should not be used in this profile. It is a reference to an object of type IpamDnsProviderProfile.
	EastWestIPAMProviderRef *string `json:"east_west_ipam_provider_ref,omitempty"`

	// Enable VIP on all data interfaces for the Cloud. Field introduced in 18.2.9, 20.1.1.
	EnableVipOnAllInterfaces *bool `json:"enable_vip_on_all_interfaces,omitempty"`

	// Use static routes for VIP side network resolution during VirtualService placement.
	EnableVipStaticRoutes *bool `json:"enable_vip_static_routes,omitempty"`

	// Google Cloud Platform Configuration. Field introduced in 18.2.1. Allowed in Basic edition, Essentials edition, Enterprise edition.
	GcpConfiguration *GCPConfiguration `json:"gcp_configuration,omitempty"`

	// Enable IPv6 auto configuration. Field introduced in 18.1.1.
	Ip6AutocfgEnabled *bool `json:"ip6_autocfg_enabled,omitempty"`

	// Ipam Profile for the cloud. It is a reference to an object of type IpamDnsProviderProfile.
	IPAMProviderRef *string `json:"ipam_provider_ref,omitempty"`

	// Specifies the default license tier which would be used by new SE Groups. This field by default inherits the value from system configuration. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS. Field introduced in 17.2.5.
	LicenseTier *string `json:"license_tier,omitempty"`

	// If no license type is specified then default license enforcement for the cloud type is chosen. The default mappings are Container Cloud is Max Ses, OpenStack and VMware is cores and linux it is Sockets. Enum options - LIC_BACKEND_SERVERS, LIC_SOCKETS, LIC_CORES, LIC_HOSTS, LIC_SE_BANDWIDTH, LIC_METERED_SE_BANDWIDTH.
	LicenseType *string `json:"license_type,omitempty"`

	// Placeholder for description of property linuxserver_configuration of obj type Cloud field type str  type object
	LinuxserverConfiguration *LinuxServerConfiguration `json:"linuxserver_configuration,omitempty"`

	//  Field deprecated in 18.2.2.
	MesosConfiguration *MesosConfiguration `json:"mesos_configuration,omitempty"`

	// MTU setting for the cloud. Unit is BYTES.
	Mtu *int32 `json:"mtu,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// Configuration parameters for NSX Manager. Field introduced in 17.1.1.
	NsxConfiguration *NsxConfiguration `json:"nsx_configuration,omitempty"`

	// NSX-T Cloud Platform Configuration. Field introduced in 20.1.1. Allowed in Essentials edition, Enterprise edition.
	NsxtConfiguration *NsxtConfiguration `json:"nsxt_configuration,omitempty"`

	// Default prefix for all automatically created objects in this cloud. This prefix can be overridden by the SE-Group template.
	ObjNamePrefix *string `json:"obj_name_prefix,omitempty"`

	// Placeholder for description of property openstack_configuration of obj type Cloud field type str  type object
	OpenstackConfiguration *OpenStackConfiguration `json:"openstack_configuration,omitempty"`

	//  Field deprecated in 20.1.1.
	Oshiftk8sConfiguration *OShiftK8SConfiguration `json:"oshiftk8s_configuration,omitempty"`

	// Prefer static routes over interface routes during VirtualService placement.
	PreferStaticRoutes *bool `json:"prefer_static_routes,omitempty"`

	// Placeholder for description of property proxy_configuration of obj type Cloud field type str  type object
	ProxyConfiguration *ProxyConfiguration `json:"proxy_configuration,omitempty"`

	// Placeholder for description of property rancher_configuration of obj type Cloud field type str  type object
	RancherConfiguration *RancherConfiguration `json:"rancher_configuration,omitempty"`

	// The Service Engine Group to use as template. It is a reference to an object of type ServiceEngineGroup. Field introduced in 18.2.5.
	SeGroupTemplateRef *string `json:"se_group_template_ref,omitempty"`

	// DNS records for VIPs are added/deleted based on the operational state of the VIPs. Field introduced in 17.1.12. Allowed in Basic(Allowed values- true) edition, Essentials(Allowed values- true) edition, Enterprise edition.
	StateBasedDNSRegistration *bool `json:"state_based_dns_registration,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Placeholder for description of property vca_configuration of obj type Cloud field type str  type object
	VcaConfiguration *VCloudAirConfiguration `json:"vca_configuration,omitempty"`

	// Placeholder for description of property vcenter_configuration of obj type Cloud field type str  type object
	VcenterConfiguration *VCenterConfiguration `json:"vcenter_configuration,omitempty"`

	// Cloud type. Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT. Allowed in Basic(Allowed values- CLOUD_NONE,CLOUD_NSXT) edition, Essentials(Allowed values- CLOUD_NONE,CLOUD_VCENTER) edition, Enterprise edition.
	// Required: true
	Vtype *string `json:"vtype"`
}
