// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HypervisorProperties hypervisor properties
// swagger:model Hypervisor_Properties
type HypervisorProperties struct {

	//  Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Htype *string `json:"htype"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxIpsPerNic uint32 `json:"max_ips_per_nic,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MaxNics uint32 `json:"max_nics,omitempty"`
}
