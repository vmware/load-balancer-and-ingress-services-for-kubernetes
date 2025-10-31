// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AzureInfo azure info
// swagger:model AzureInfo
type AzureInfo struct {

	// Name of the availability set of which the VM is a part of. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AvailabilitySet *string `json:"availability_set,omitempty"`

	// Fault domain within the availability set the VM is a part of. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FaultDomain *string `json:"fault_domain,omitempty"`

	// Name of the Azure VM. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`

	// Resource group name for the VM. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResourceGroup *string `json:"resource_group,omitempty"`

	// Subnet ID of the primary nic of the VM. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SubnetID *string `json:"subnet_id,omitempty"`

	// Update domain within the availability set the VM is a part of. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UpdateDomain *string `json:"update_domain,omitempty"`

	// Azure VM uuid for the SE VM. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VMUUID *string `json:"vm_uuid,omitempty"`

	// VNIC id of the primary nic of the VM. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	VnicID *string `json:"vnic_id,omitempty"`
}
