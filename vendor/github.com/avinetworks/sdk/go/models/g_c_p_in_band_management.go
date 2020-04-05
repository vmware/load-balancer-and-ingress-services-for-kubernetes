package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GCPInBandManagement g c p in band management
// swagger:model GCPInBandManagement
type GCPInBandManagement struct {

	// Service Engine Network Name. Field introduced in 18.2.2.
	// Required: true
	VpcNetworkName *string `json:"vpc_network_name"`

	// Project ID of the Service Engine Network. By default, Service Engine Project ID will be used. Field introduced in 18.2.1.
	VpcProjectID *string `json:"vpc_project_id,omitempty"`

	// Service Engine Network Subnet Name. Field introduced in 18.2.1.
	// Required: true
	VpcSubnetName *string `json:"vpc_subnet_name"`
}
