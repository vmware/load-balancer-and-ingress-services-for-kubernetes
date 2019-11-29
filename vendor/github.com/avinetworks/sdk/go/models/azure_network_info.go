package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AzureNetworkInfo azure network info
// swagger:model AzureNetworkInfo
type AzureNetworkInfo struct {

	// Id of the Azure subnet where Avi Controller will create the Service Engines. . Field introduced in 17.2.1.
	SeNetworkID *string `json:"se_network_id,omitempty"`

	// Virtual network where Virtual IPs will belong. Field introduced in 17.2.1.
	VirtualNetworkID *string `json:"virtual_network_id,omitempty"`
}
