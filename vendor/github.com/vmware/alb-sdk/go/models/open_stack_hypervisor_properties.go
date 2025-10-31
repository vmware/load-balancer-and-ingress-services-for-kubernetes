// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OpenStackHypervisorProperties open stack hypervisor properties
// swagger:model OpenStackHypervisorProperties
type OpenStackHypervisorProperties struct {

	// Hypervisor type. Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Hypervisor *string `json:"hypervisor"`

	// Custom properties to be associated with the SE image in Glance for this hypervisor type. Field introduced in 17.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ImageProperties []*Property `json:"image_properties,omitempty"`
}
