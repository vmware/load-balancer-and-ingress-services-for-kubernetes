package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AzureSetup azure setup
// swagger:model AzureSetup
type AzureSetup struct {

	// alb_id of AzureSetup.
	AlbID *string `json:"alb_id,omitempty"`

	// cc_id of AzureSetup.
	CcID *string `json:"cc_id,omitempty"`

	// nic_id of AzureSetup.
	NicID *string `json:"nic_id,omitempty"`

	// reason of AzureSetup.
	Reason *string `json:"reason,omitempty"`

	// resource_group of AzureSetup.
	ResourceGroup *string `json:"resource_group,omitempty"`

	// status of AzureSetup.
	Status *string `json:"status,omitempty"`

	// subscription_id of AzureSetup.
	SubscriptionID *string `json:"subscription_id,omitempty"`

	// Placeholder for description of property vips of obj type AzureSetup field type str  type object
	Vips []*IPAddr `json:"vips,omitempty"`

	// vnet_id of AzureSetup.
	VnetID *string `json:"vnet_id,omitempty"`

	// Unique object identifiers of vss.
	VsUuids []string `json:"vs_uuids,omitempty"`
}
