package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AzureMarketplace azure marketplace
// swagger:model AzureMarketplace
type AzureMarketplace struct {

	// Azure Cloud id. Field introduced in 18.2.2.
	CcID *string `json:"cc_id,omitempty"`

	// Avi azure marketplace offer name. Field introduced in 18.2.2.
	Offer *string `json:"offer,omitempty"`

	// Avi azure marketplace publisher name. Field introduced in 18.2.2.
	Publisher *string `json:"publisher,omitempty"`

	// Azure marketplace license term failure status. Field introduced in 18.2.2.
	Reason *string `json:"reason,omitempty"`

	// Azure resource group name. Field introduced in 18.2.2.
	ResourceGroup *string `json:"resource_group,omitempty"`

	// Avi azure marketplace skus list. Field introduced in 18.2.2.
	Skus []string `json:"skus,omitempty"`

	// Azure marketplace license term acceptance status. Field introduced in 18.2.2.
	Status *string `json:"status,omitempty"`

	// Azure Subscription id. Field introduced in 18.2.2.
	SubscriptionID *string `json:"subscription_id,omitempty"`

	// Azure Vnet id. Field introduced in 18.2.2.
	VnetID *string `json:"vnet_id,omitempty"`
}
