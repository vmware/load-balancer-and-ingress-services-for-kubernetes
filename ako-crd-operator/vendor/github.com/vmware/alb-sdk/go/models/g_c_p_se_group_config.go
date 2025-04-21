// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GCPSeGroupConfig g c p se group config
// swagger:model GCPSeGroupConfig
type GCPSeGroupConfig struct {

	// Service Engine Backend Data Network Name, used only for GCP cloud.Overrides the cloud level setting for Backend Data Network in GCP Two Arm Mode. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BackendDataVpcNetworkName *string `json:"backend_data_vpc_network_name,omitempty"`

	// Project ID of the Service Engine Backend Data Network. By default, Service Engine Project ID will be used. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BackendDataVpcProjectID *string `json:"backend_data_vpc_project_id,omitempty"`

	// Service Engine Backend Data Subnet Name, used only for GCP cloud.Overrides the cloud level setting for Backend Data Subnet in GCP Two Arm Mode. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	BackendDataVpcSubnetName *string `json:"backend_data_vpc_subnet_name,omitempty"`
}
