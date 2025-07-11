// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPAMDNSAzureProfile ipam Dns azure profile
// swagger:model IpamDnsAzureProfile
type IPAMDNSAzureProfile struct {

	// Service principal based credentials for azure. Only one of azure_userpass or azure_serviceprincipal is allowed. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AzureServiceprincipal *AzureServicePrincipalCredentials `json:"azure_serviceprincipal,omitempty"`

	// User name password based credentials for azure. Only one of azure_userpass or azure_serviceprincipal is allowed. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AzureUserpass *AzureUserPassCredentials `json:"azure_userpass,omitempty"`

	// Used for allocating egress service source IPs. Field introduced in 17.2.8. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EgressServiceSubnets []string `json:"egress_service_subnets,omitempty"`

	// Azure resource group dedicated for Avi Controller. Avi Controller will create all its resources in this resource group. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceGroup *string `json:"resource_group,omitempty"`

	// Subscription Id for the Azure subscription. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SubscriptionID *string `json:"subscription_id,omitempty"`

	// Usable domains to pick from Azure DNS. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UsableDomains []string `json:"usable_domains,omitempty"`

	// Usable networks for Virtual IP. If VirtualService does not specify a network and auto_allocate_ip is set, then the first available network from this list will be chosen for IP allocation. Field introduced in 17.2.1. Maximum of 128 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UsableNetworkUuids []string `json:"usable_network_uuids,omitempty"`

	// Use Azure's enhanced HA features. This needs a public IP to be associated with the VIP. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseEnhancedHa *bool `json:"use_enhanced_ha,omitempty"`

	// Use Standard SKU Azure Load Balancer. By default Basic SKU Load Balancer is used. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseStandardAlb *bool `json:"use_standard_alb,omitempty"`

	// Virtual networks where Virtual IPs will belong. Field introduced in 17.2.1. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VirtualNetworkIds []string `json:"virtual_network_ids,omitempty"`
}
