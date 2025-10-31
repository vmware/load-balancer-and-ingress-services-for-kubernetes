// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VcenterHosts vcenter hosts
// swagger:model VcenterHosts
type VcenterHosts struct {

	//  It is a reference to an object of type VIMgrHostRuntime. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostRefs []string `json:"host_refs,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Include *bool `json:"include,omitempty"`
}
