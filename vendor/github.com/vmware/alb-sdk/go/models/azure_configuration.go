// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AzureConfiguration azure configuration
// swagger:model AzureConfiguration
type AzureConfiguration struct {

	// Availability zones to be used in Azure. Field introduced in 17.2.5. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvailabilityZones []string `json:"availability_zones,omitempty"`

	// Credentials to access azure cloud. It is a reference to an object of type CloudConnectorUser. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloudCredentialsRef *string `json:"cloud_credentials_ref,omitempty"`

	// Disks Encryption Set resource-id (des_id) to encrypt se image and managed disk using customer-managed-keys. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DesID *string `json:"des_id,omitempty"`

	// Azure location where this cloud will be located. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Location *string `json:"location,omitempty"`

	// Azure virtual network and subnet information. Field introduced in 17.2.1. Minimum of 1 items required. Maximum of 1 items allowed. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NetworkInfo []*AzureNetworkInfo `json:"network_info,omitempty"`

	// Azure resource group dedicated for Avi Controller. Avi Controller will create all its resources in this resource group. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceGroup *string `json:"resource_group,omitempty"`

	// Storage Account to be used for uploading SE VHD images to Azure. Must include the resource group name. Format '<resource-group> <storage-account-name>'. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SeStorageAccount *string `json:"se_storage_account,omitempty"`

	// Subscription Id for the Azure subscription. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SubscriptionID *string `json:"subscription_id,omitempty"`

	// Azure is the DNS provider. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseAzureDNS *bool `json:"use_azure_dns,omitempty"`

	// Use Azure's enhanced HA features. This needs a public IP to be associated with the VIP. . Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseEnhancedHa *bool `json:"use_enhanced_ha,omitempty"`

	// Use Azure managed disks for SE storage. Field introduced in 17.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseManagedDisks *bool `json:"use_managed_disks,omitempty"`

	// Use Standard SKU Azure Load Balancer. By default Standard SKU Load Balancer is used. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UseStandardAlb *bool `json:"use_standard_alb,omitempty"`
}
