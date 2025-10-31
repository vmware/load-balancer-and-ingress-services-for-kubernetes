// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudConfig cloud config
// swagger:model CloudConfig
type CloudConfig struct {

	// CloudConnector polling interval in seconds for external autoscale groups, minimum 60 seconds. Allowed values are 60-3600. Field introduced in 22.1.1. Unit is SECONDS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AutoscalePollingInterval *uint32 `json:"autoscale_polling_interval,omitempty"`

	// Select the IP address management scheme. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DhcpEnabled *bool `json:"dhcp_enabled,omitempty"`

	//  Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DNSResolutionOnSe *bool `json:"dns_resolution_on_se,omitempty"`

	// Enable VIP on all data interfaces for the Cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableVipOnAllInterfaces *bool `json:"enable_vip_on_all_interfaces,omitempty"`

	// Use static routes for VIP side network resolution during VirtualService placement. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EnableVipStaticRoutes *bool `json:"enable_vip_static_routes,omitempty"`

	// Enable IPv6 auto configuration. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Ip6AutocfgEnabled *bool `json:"ip6_autocfg_enabled,omitempty"`

	// Specifies the default license tier which would be used by new SE Groups. This field by default inherits the value from system configuration. Enum options - ENTERPRISE_16, ENTERPRISE, ENTERPRISE_18, BASIC, ESSENTIALS, ENTERPRISE_WITH_CLOUD_SERVICES. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LicenseTier *string `json:"license_tier,omitempty"`

	// If no license type is specified then default license enforcement for the cloud type is chosen. The default mappings are Container Cloud is Max Ses, OpenStack and VMware is cores and linux it is Sockets. Enum options - LIC_BACKEND_SERVERS, LIC_SOCKETS, LIC_CORES, LIC_HOSTS, LIC_SE_BANDWIDTH, LIC_METERED_SE_BANDWIDTH. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LicenseType *string `json:"license_type,omitempty"`

	// Cloud is in maintenance mode. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaintenanceMode *bool `json:"maintenance_mode,omitempty"`

	// MTU setting for the cloud. Field introduced in 22.1.1. Unit is BYTES. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Mtu *uint32 `json:"mtu,omitempty"`

	// Name of the cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Prefer static routes over interface routes during VirtualService placement. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	PreferStaticRoutes *bool `json:"prefer_static_routes,omitempty"`

	// DNS records for VIPs are added/deleted based on the operational state of the VIPs. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StateBasedDNSRegistration *bool `json:"state_based_dns_registration,omitempty"`

	//  It is a reference to an object of type Tenant. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// URL of the cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	URL *string `json:"url,omitempty"`

	// UUID of the cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`

	// VCenter configuration of the cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VcenterConfiguration *VCenterConfiguration `json:"vcenter_configuration,omitempty"`

	// This deployment is VMware on AWS cloud. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VmcDeployment *bool `json:"vmc_deployment,omitempty"`

	// Cloud type. Enum options - CLOUD_NONE, CLOUD_VCENTER, CLOUD_OPENSTACK, CLOUD_AWS, CLOUD_VCA, CLOUD_APIC, CLOUD_MESOS, CLOUD_LINUXSERVER, CLOUD_DOCKER_UCP, CLOUD_RANCHER, CLOUD_OSHIFT_K8S, CLOUD_AZURE, CLOUD_GCP, CLOUD_NSXT. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Vtype *string `json:"vtype"`
}
