package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// IPAMDNSGCPProfile ipam Dns g c p profile
// swagger:model IpamDnsGCPProfile
type IPAMDNSGCPProfile struct {

	// Match SE group subnets for VIP placement. Default is to not match SE group subnets. Field introduced in 17.1.1.
	MatchSeGroupSubnet *bool `json:"match_se_group_subnet,omitempty"`

	// Google Cloud Platform Network Host Project ID. This is the host project in which Google Cloud Platform Network resides. Field introduced in 18.1.2.
	NetworkHostProjectID *string `json:"network_host_project_id,omitempty"`

	// Google Cloud Platform Region Name. Field introduced in 18.1.2.
	RegionName *string `json:"region_name,omitempty"`

	// Google Cloud Platform Project ID. This is the project where service engines are hosted. This field is optional. By default it will use the value of the field network_host_project_id. Field introduced in 18.1.2.
	SeProjectID *string `json:"se_project_id,omitempty"`

	// Usable networks for Virtual IP. If VirtualService does not specify a network and auto_allocate_ip is set, then the first available network from this list will be chosen for IP allocation. It is a reference to an object of type Network.
	UsableNetworkRefs []string `json:"usable_network_refs,omitempty"`

	// Use Google Cloud Platform Network for Private VIP allocation. By default Avi Vantage Network is used for Private VIP allocation. Field introduced in 18.1.2.
	UseGcpNetwork *bool `json:"use_gcp_network,omitempty"`

	// Google Cloud Platform VPC Network Name. Field introduced in 18.1.2.
	VpcNetworkName *string `json:"vpc_network_name,omitempty"`
}
