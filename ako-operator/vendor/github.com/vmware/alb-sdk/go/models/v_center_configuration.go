// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VCenterConfiguration v center configuration
// swagger:model vCenterConfiguration
type VCenterConfiguration struct {

	// vCenter content library where Service Engine images are stored. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ContentLib *ContentLibConfig `json:"content_lib,omitempty"`

	// Datacenter for virtual infrastructure discovery. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Datacenter *string `json:"datacenter,omitempty"`

	// Managed object id of the datacenter. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DatacenterManagedObjectID *string `json:"datacenter_managed_object_id,omitempty"`

	// If true, VM's on the vCenter will not be discovered.Set it to true if there are more than 10000 VMs in the datacenter. Field deprecated in 30.1.1. Field introduced in 20.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DeactivateVMDiscovery *bool `json:"deactivate_vm_discovery,omitempty"`

	// If true, NSX-T segment spanning multiple VDS with vCenter cloud are merged to a single network in Avi. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IsNsxEnvironment *bool `json:"is_nsx_environment,omitempty"`

	// Management subnet to use for Avi Service Engines. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ManagementIPSubnet *IPAddrPrefix `json:"management_ip_subnet,omitempty"`

	// Management network to use for Avi Service Engines. It is a reference to an object of type VIMgrNWRuntime. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ManagementNetwork *string `json:"management_network,omitempty"`

	// The password Avi Vantage will use when authenticating with vCenter. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Password *string `json:"password,omitempty"`

	// Set the access mode to vCenter as either Read, which allows Avi to discover networks and servers, or Write, which also allows Avi to create Service Engines and configure their network properties. Enum options - NO_ACCESS, READ_ACCESS, WRITE_ACCESS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Privilege *string `json:"privilege"`

	// If false, Service Engine image will not be pushed to content library. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Essentials edition with any value, Enterprise with Cloud Services edition.
	UseContentLib *bool `json:"use_content_lib,omitempty"`

	// The username Avi Vantage will use when authenticating with vCenter. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Username *string `json:"username,omitempty"`

	// Avi Service Engine Template in vCenter to be used for creating Service Engines. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterTemplateSeLocation *string `json:"vcenter_template_se_location,omitempty"`

	// vCenter hostname or IP address. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VcenterURL *string `json:"vcenter_url,omitempty"`
}
