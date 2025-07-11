// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPTwoArmMode g c p two arm mode
// swagger:model GCPTwoArmMode
type GCPTwoArmMode struct {

	// Service Engine Backend Data Network Name. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	BackendDataVpcNetworkName *string `json:"backend_data_vpc_network_name"`

	// Project ID of the Service Engine Backend Data Network. By default, Service Engine Project ID will be used. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BackendDataVpcProjectID *string `json:"backend_data_vpc_project_id,omitempty"`

	// Service Engine Backend Data Network Subnet Name. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	BackendDataVpcSubnetName *string `json:"backend_data_vpc_subnet_name"`

	// Service Engine Frontend Data Network Name. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	FrontendDataVpcNetworkName *string `json:"frontend_data_vpc_network_name"`

	// Project ID of the Service Engine Frontend Data Network. By default, Service Engine Project ID will be used. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FrontendDataVpcProjectID *string `json:"frontend_data_vpc_project_id,omitempty"`

	// Service Engine Frontend Data Network Subnet Name. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	FrontendDataVpcSubnetName *string `json:"frontend_data_vpc_subnet_name"`

	// Service Engine Management Network Name. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ManagementVpcNetworkName *string `json:"management_vpc_network_name"`

	// Project ID of the Service Engine Management Network. By default, Service Engine Project ID will be used. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ManagementVpcProjectID *string `json:"management_vpc_project_id,omitempty"`

	// Service Engine Management Network Subnet Name. Field introduced in 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	ManagementVpcSubnetName *string `json:"management_vpc_subnet_name"`
}
