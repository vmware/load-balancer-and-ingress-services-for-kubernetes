package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// TwoArmMode two arm mode
// swagger:model TwoArmMode
type TwoArmMode struct {

	// Service Engine Backend Data Network Subnet Name. Field introduced in 18.2.1.
	// Required: true
	BackendDataVpcSubnetName *string `json:"backend_data_vpc_subnet_name"`

	// Project ID of the Service Engine Frontend Data Network. Field introduced in 18.2.1.
	FrontendDataVpcProjectID *string `json:"frontend_data_vpc_project_id,omitempty"`

	// Service Engine Frontend Data Network Subnet Name. Field introduced in 18.2.1.
	// Required: true
	FrontendDataVpcSubnetName *string `json:"frontend_data_vpc_subnet_name"`

	// Service Engine Management Network Subnet Name. Field introduced in 18.2.1.
	// Required: true
	ManagementVpcSubnetName *string `json:"management_vpc_subnet_name"`
}
