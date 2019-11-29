package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AzureConfiguration azure configuration
// swagger:model AzureConfiguration
type AzureConfiguration struct {

	// Availability zones to be used in Azure. Field introduced in 17.2.5.
	AvailabilityZones []string `json:"availability_zones,omitempty"`

	// Credentials to access azure cloud. It is a reference to an object of type CloudConnectorUser. Field introduced in 17.2.1.
	CloudCredentialsRef *string `json:"cloud_credentials_ref,omitempty"`

	// Azure location where this cloud will be located. Field introduced in 17.2.1.
	Location *string `json:"location,omitempty"`

	// Azure virtual network and subnet information. Field introduced in 17.2.1.
	NetworkInfo []*AzureNetworkInfo `json:"network_info,omitempty"`

	// Azure resource group dedicated for Avi Controller. Avi Controller will create all its resources in this resource group. Field introduced in 17.2.1.
	ResourceGroup *string `json:"resource_group,omitempty"`

	// Subscription Id for the Azure subscription. Field introduced in 17.2.1.
	SubscriptionID *string `json:"subscription_id,omitempty"`

	// Azure is the DNS provider. Field introduced in 17.2.1.
	UseAzureDNS *bool `json:"use_azure_dns,omitempty"`

	// Use Azure's enhanced HA features. This needs a public IP to be associated with the VIP. . Field introduced in 17.2.1.
	UseEnhancedHa *bool `json:"use_enhanced_ha,omitempty"`

	// Use Azure managed disks for SE storage. Field introduced in 17.2.2.
	UseManagedDisks *bool `json:"use_managed_disks,omitempty"`

	// Use Standard SKU Azure Load Balancer. By default Basic SKU Load Balancer is used. Field introduced in 17.2.7.
	UseStandardAlb *bool `json:"use_standard_alb,omitempty"`
}
