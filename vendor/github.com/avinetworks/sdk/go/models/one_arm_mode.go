package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OneArmMode one arm mode
// swagger:model OneArmMode
type OneArmMode struct {

	// Project ID of the Service Engine Data Network. Field introduced in 18.2.1.
	DataVpcProjectID *string `json:"data_vpc_project_id,omitempty"`

	// Service Engine Data Network Subnet Name. Field introduced in 18.2.1.
	// Required: true
	DataVpcSubnetName *string `json:"data_vpc_subnet_name"`

	// Service Engine Management Network Subnet Name. Field introduced in 18.2.1.
	// Required: true
	ManagementVpcSubnetName *string `json:"management_vpc_subnet_name"`
}
