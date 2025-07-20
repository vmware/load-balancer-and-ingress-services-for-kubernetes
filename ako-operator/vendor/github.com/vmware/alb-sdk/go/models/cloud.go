// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Cloud cloud
// swagger:model Cloud
type Cloud struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// CloudConnector polling interval in seconds for external autoscale groups, minimum 60 seconds. Allowed values are 60-3600. Field introduced in 18.2.2. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 60), Basic edition(Allowed values- 60), Enterprise with Cloud Services edition.
	AutoscalePollingInterval *uint32 `json:"autoscale_polling_interval,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AwsConfiguration *AwsConfiguration `json:"aws_configuration,omitempty"`

	//  Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AzureConfiguration *AzureConfiguration `json:"azure_configuration,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CloudstackConfiguration *CloudStackConfiguration `json:"cloudstack_configuration,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Custom tags for all Avi created resources in the cloud infrastructure. Field introduced in 17.1.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CustomTags []*CustomTag `json:"custom_tags,omitempty"`

	// Select the IP address management scheme. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DhcpEnabled *bool `json:"dhcp_enabled,omitempty"`

	// DNS Profile for the cloud. It is a reference to an object of type IpamDnsProviderProfile. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSProviderRef *string `json:"dns_provider_ref,omitempty"`

	// By default, pool member FQDNs are resolved on the Controller. When this is set, pool member FQDNs are instead resolved on Service Engines in this cloud. This is useful in scenarios where pool member FQDNs can only be resolved from Service Engines and not from the Controller. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- false), Basic edition(Allowed values- false), Enterprise with Cloud Services edition.
	DNSResolutionOnSe *bool `json:"dns_resolution_on_se,omitempty"`

	// DNS resolver for the cloud. Field introduced in 20.1.5. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSResolvers []*DNSResolver `json:"dns_resolvers,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DockerConfiguration *DockerConfiguration `json:"docker_configuration,omitempty"`

	// DNS Profile for East-West services. It is a reference to an object of type IpamDnsProviderProfile. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EastWestDNSProviderRef *string `json:"east_west_dns_provider_ref,omitempty"`

	// Ipam Profile for East-West services. Warning - Please use virtual subnets in this IPAM profile that do not conflict with the underlay networks or any overlay networks in the cluster. For example in AWS and GCP, 169.254.0.0/16 is used for storing instance metadata. Hence, it should not be used in this profile. It is a reference to an object of type IpamDnsProviderProfile. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EastWestIPAMProviderRef *string `json:"east_west_ipam_provider_ref,omitempty"`

	// Enable VIP on all data interfaces for the Cloud. Field introduced in 18.2.9, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableVipOnAllInterfaces *bool `json:"enable_vip_on_all_interfaces,omitempty"`

	// Use static routes for VIP side network resolution during VirtualService placement. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableVipStaticRoutes *bool `json:"enable_vip_static_routes,omitempty"`

	// Google Cloud Platform Configuration. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	GcpConfiguration *GCPConfiguration `json:"gcp_configuration,omitempty"`

	// Enable IPv6 auto configuration. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Ip6AutocfgEnabled *bool `json:"ip6_autocfg_enabled,omitempty"`

	// Ipam Profile for the cloud. It is a reference to an object of type IpamDnsProviderProfile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPAMProviderRef *string `json:"ipam_provider_ref,omitempty"`

	// Specifies the default license tier which would be used by new SE Groups. This field by default inherits the value from system configuration. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS, ENTERPRISE_WITH_CLOUD_SERVICES. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseTier *string `json:"license_tier,omitempty"`

	// If no license type is specified then default license enforcement for the cloud type is chosen. The default mappings are Container Cloud is Max Ses, OpenStack and VMware is cores and linux it is Sockets. Enum options - LIC_BACKEND_SERVERS, LIC_SOCKETS, LIC_CORES, LIC_HOSTS, LIC_SE_BANDWIDTH, LIC_METERED_SE_BANDWIDTH. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	LicenseType *string `json:"license_type,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LinuxserverConfiguration *LinuxServerConfiguration `json:"linuxserver_configuration,omitempty"`

	// Cloud is in maintenance mode. Field introduced in 20.1.7,21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaintenanceMode *bool `json:"maintenance_mode,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Cloud metrics collector polling interval in seconds. Field introduced in 22.1.1. Unit is SECONDS. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	MetricsPollingInterval *uint32 `json:"metrics_polling_interval,omitempty"`

	// Enable IPv4 on the Management interface of the ServiceEngine. Defaults to dhcp if no static config on Network present. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MgmtIPV4Enabled *bool `json:"mgmt_ip_v4_enabled,omitempty"`

	// Enable IPv6 on the Management interface of the ServiceEngine. Defaults to autocfg if no static config on Network present. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MgmtIPV6Enabled *bool `json:"mgmt_ip_v6_enabled,omitempty"`

	// MTU setting for the cloud. Unit is BYTES. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Mtu *uint32 `json:"mtu,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// NSX-T Cloud Platform Configuration. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Basic, Enterprise with Cloud Services edition.
	NsxtConfiguration *NsxtConfiguration `json:"nsxt_configuration,omitempty"`

	// NTP Configuration for the cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NtpConfiguration *NTPConfiguration `json:"ntp_configuration,omitempty"`

	// Default prefix for all automatically created objects in this cloud. This prefix can be overridden by the SE-Group template. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ObjNamePrefix *string `json:"obj_name_prefix,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	OpenstackConfiguration *OpenStackConfiguration `json:"openstack_configuration,omitempty"`

	// Prefer static routes over interface routes during VirtualService placement. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PreferStaticRoutes *bool `json:"prefer_static_routes,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ProxyConfiguration *ProxyConfiguration `json:"proxy_configuration,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RancherConfiguration *RancherConfiguration `json:"rancher_configuration,omitempty"`

	// Resolve IPv6 address for pool member FQDNs. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ResolveFqdnToIPV6 *bool `json:"resolve_fqdn_to_ipv6,omitempty"`

	// The Service Engine Group to use as template. It is a reference to an object of type ServiceEngineGroup. Field introduced in 18.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeGroupTemplateRef *string `json:"se_group_template_ref,omitempty"`

	// DNS records for VIPs are added/deleted based on the operational state of the VIPs. Field introduced in 17.1.12. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- true), Basic edition(Allowed values- true), Enterprise with Cloud Services edition.
	StateBasedDNSRegistration *bool `json:"state_based_dns_registration,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	//  Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcaConfiguration *VCloudAirConfiguration `json:"vca_configuration,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Enterprise with Cloud Services edition.
	VcenterConfiguration *VCenterConfiguration `json:"vcenter_configuration,omitempty"`

	// This deployment is VMware on AWS cloud. Field introduced in 20.1.5, 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VmcDeployment *bool `json:"vmc_deployment,omitempty"`

	// Cloud type. Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- CLOUD_NONE,CLOUD_VCENTER), Basic edition(Allowed values- CLOUD_NONE,CLOUD_NSXT), Enterprise with Cloud Services edition.
	// Required: true
	Vtype *string `json:"vtype"`
}
