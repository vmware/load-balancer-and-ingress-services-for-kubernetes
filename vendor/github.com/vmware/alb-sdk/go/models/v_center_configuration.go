// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VCenterConfiguration v center configuration
// swagger:model vCenterConfiguration
type VCenterConfiguration struct {

	// Datacenter for virtual infrastructure discovery.
	Datacenter *string `json:"datacenter,omitempty"`

	// If true, VM's on the vCenter will not be discovered.Set it to true if there are more than 10000 VMs in the datacenter. Field introduced in 20.1.5.
	DeactivateVMDiscovery *bool `json:"deactivate_vm_discovery,omitempty"`

	// Management subnet to use for Avi Service Engines.
	ManagementIPSubnet *IPAddrPrefix `json:"management_ip_subnet,omitempty"`

	// Management network to use for Avi Service Engines. It is a reference to an object of type VIMgrNWRuntime.
	ManagementNetwork *string `json:"management_network,omitempty"`

	// The password Avi Vantage will use when authenticating with vCenter.
	Password *string `json:"password,omitempty"`

	// Set the access mode to vCenter as either Read, which allows Avi to discover networks and servers, or Write, which also allows Avi to create Service Engines and configure their network properties. Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS.
	// Required: true
	Privilege *string `json:"privilege"`

	// The username Avi Vantage will use when authenticating with vCenter.
	Username *string `json:"username,omitempty"`

	// Avi Service Engine Template in vCenter to be used for creating Service Engines.
	VcenterTemplateSeLocation *string `json:"vcenter_template_se_location,omitempty"`

	// vCenter hostname or IP address.
	VcenterURL *string `json:"vcenter_url,omitempty"`
}
