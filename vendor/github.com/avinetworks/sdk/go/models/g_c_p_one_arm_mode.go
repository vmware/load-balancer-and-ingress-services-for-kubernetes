package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GCPOneArmMode g c p one arm mode
// swagger:model GCPOneArmMode
type GCPOneArmMode struct {

	// Service Engine Data Network Name. Field introduced in 18.2.2.
	// Required: true
	DataVpcNetworkName *string `json:"data_vpc_network_name"`

	// Project ID of the Service Engine Data Network. By default, Service Engine Project ID will be used. Field introduced in 18.2.1.
	DataVpcProjectID *string `json:"data_vpc_project_id,omitempty"`

	// Service Engine Data Network Subnet Name. Field introduced in 18.2.1.
	// Required: true
	DataVpcSubnetName *string `json:"data_vpc_subnet_name"`

	// Service Engine Management Network Name. Field introduced in 18.2.2.
	// Required: true
	ManagementVpcNetworkName *string `json:"management_vpc_network_name"`

	// Service Engine Management Network Subnet Name. Field introduced in 18.2.1.
	// Required: true
	ManagementVpcSubnetName *string `json:"management_vpc_subnet_name"`
}
