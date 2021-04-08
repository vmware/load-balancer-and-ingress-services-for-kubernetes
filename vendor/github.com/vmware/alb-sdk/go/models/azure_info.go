package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AzureInfo azure info
// swagger:model AzureInfo
type AzureInfo struct {

	// Name of the availability set of which the VM is a part of. Field introduced in 17.2.1.
	AvailabilitySet *string `json:"availability_set,omitempty"`

	// Fault domain within the availability set the VM is a part of. Field introduced in 17.2.1.
	FaultDomain *string `json:"fault_domain,omitempty"`

	// Name of the Azure VM. Field introduced in 17.2.1.
	Name *string `json:"name,omitempty"`

	// Resource group name for the VM. Field introduced in 17.2.1.
	ResourceGroup *string `json:"resource_group,omitempty"`

	// Subnet ID of the primary nic of the VM. Field introduced in 17.2.1.
	SubnetID *string `json:"subnet_id,omitempty"`

	// Update domain within the availability set the VM is a part of. Field introduced in 17.2.1.
	UpdateDomain *string `json:"update_domain,omitempty"`

	// Azure VM uuid for the SE VM. Field introduced in 17.2.1.
	VMUUID *string `json:"vm_uuid,omitempty"`

	// VNIC id of the primary nic of the VM. Field introduced in 17.2.1.
	VnicID *string `json:"vnic_id,omitempty"`
}
