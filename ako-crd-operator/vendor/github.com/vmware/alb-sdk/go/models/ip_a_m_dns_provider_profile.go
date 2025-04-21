// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMDNSProviderProfile ipam Dns provider profile
// swagger:model IpamDnsProviderProfile
type IPAMDNSProviderProfile struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// If this flag is set, only allocate IP from networks in the Virtual Service VRF. Applicable for Avi Vantage IPAM only. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AllocateIPInVrf *bool `json:"allocate_ip_in_vrf,omitempty"`

	// Provider details if type is AWS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AwsProfile *IPAMDNSAwsProfile `json:"aws_profile,omitempty"`

	// Provider details if type is Microsoft Azure. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AzureProfile *IPAMDNSAzureProfile `json:"azure_profile,omitempty"`

	// Protobuf versioning for config pbs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	ConfigpbAttributes *ConfigPbAttributes `json:"configpb_attributes,omitempty"`

	// Provider details if type is Custom. Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CustomProfile *IPAMDNSCustomProfile `json:"custom_profile,omitempty"`

	// Provider details if type is Google Cloud. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GcpProfile *IPAMDNSGCPProfile `json:"gcp_profile,omitempty"`

	// Provider details if type is Infoblox. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InfobloxProfile *IPAMDNSInfobloxProfile `json:"infoblox_profile,omitempty"`

	// Provider details if type is Avi. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	InternalProfile *IPAMDNSInternalProfile `json:"internal_profile,omitempty"`

	// List of labels to be used for granular RBAC. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Essentials edition with any value, Basic edition with any value, Enterprise with Cloud Services edition.
	Markers []*RoleFilterMatchLabel `json:"markers,omitempty"`

	// Name for the IPAM/DNS Provider profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`

	// Provider details for Oracle Cloud. Field introduced in 18.2.1,18.1.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OciProfile *IPAMDNSOCIprofile `json:"oci_profile,omitempty"`

	// Provider details if type is OpenStack. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	OpenstackProfile *IPAMDNSOpenstackProfile `json:"openstack_profile,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ProxyConfiguration *ProxyConfiguration `json:"proxy_configuration,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Provider details for Tencent Cloud. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TencentProfile *IPAMDNSTencentProfile `json:"tencent_profile,omitempty"`

	// Provider Type for the IPAM/DNS Provider profile. Enum options - IPAMDNS_TYPE_INFOBLOX, IPAMDNS_TYPE_AWS, IPAMDNS_TYPE_OPENSTACK, IPAMDNS_TYPE_GCP, IPAMDNS_TYPE_INFOBLOX_DNS, IPAMDNS_TYPE_CUSTOM, IPAMDNS_TYPE_CUSTOM_DNS, IPAMDNS_TYPE_AZURE, IPAMDNS_TYPE_OCI, IPAMDNS_TYPE_TENCENT, IPAMDNS_TYPE_INTERNAL, IPAMDNS_TYPE_INTERNAL_DNS, IPAMDNS_TYPE_AWS_DNS, IPAMDNS_TYPE_AZURE_DNS. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- IPAMDNS_TYPE_INTERNAL), Basic edition(Allowed values- IPAMDNS_TYPE_INTERNAL), Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID of the IPAM/DNS Provider profile. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UUID *string `json:"uuid,omitempty"`
}
