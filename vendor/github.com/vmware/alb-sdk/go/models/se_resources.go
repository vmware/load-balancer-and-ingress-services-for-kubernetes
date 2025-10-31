// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeResources se resources
// swagger:model SeResources
type SeResources struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CoresPerSocket *int32 `json:"cores_per_socket,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Disk *int32 `json:"disk"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HyperThreading *bool `json:"hyper_threading,omitempty"`

	// Indicates that the SE is running on a Virtual Machine. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HypervisorMode *bool `json:"hypervisor_mode,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Memory *int32 `json:"memory"`

	// Indicates the number of active datapath processes. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NumDatapathProcesses *uint32 `json:"num_datapath_processes,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	NumVcpus *int32 `json:"num_vcpus"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Sockets *int32 `json:"sockets,omitempty"`
}
