package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GCPSeGroupConfig g c p se group config
// swagger:model GCPSeGroupConfig
type GCPSeGroupConfig struct {

	// Service Engine Backend Data Network Name, used only for GCP cloud.Overrides the cloud level setting for Backend Data Network in GCP Two Arm Mode. Field introduced in 20.1.3.
	BackendDataVpcNetworkName *string `json:"backend_data_vpc_network_name,omitempty"`

	// Service Engine Backend Data Subnet Name, used only for GCP cloud.Overrides the cloud level setting for Backend Data Subnet in GCP Two Arm Mode. Field introduced in 20.1.3.
	BackendDataVpcSubnetName *string `json:"backend_data_vpc_subnet_name,omitempty"`
}
