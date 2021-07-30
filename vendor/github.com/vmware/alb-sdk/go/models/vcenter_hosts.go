// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VcenterHosts vcenter hosts
// swagger:model VcenterHosts
type VcenterHosts struct {

	//  It is a reference to an object of type VIMgrHostRuntime.
	HostRefs []string `json:"host_refs,omitempty"`

	// Placeholder for description of property include of obj type VcenterHosts field type str  type boolean
	Include *bool `json:"include,omitempty"`
}
