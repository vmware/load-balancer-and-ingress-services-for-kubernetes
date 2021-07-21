// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HypervisorProperties hypervisor properties
// swagger:model Hypervisor_Properties
type HypervisorProperties struct {

	//  Enum options - DEFAULT, VMWARE_ESX, KVM, VMWARE_VSAN, XEN.
	// Required: true
	Htype *string `json:"htype"`

	// Number of max_ips_per_nic.
	MaxIpsPerNic *int32 `json:"max_ips_per_nic,omitempty"`

	// Number of max_nics.
	MaxNics *int32 `json:"max_nics,omitempty"`
}
