// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// InventoryFaultConfig inventory fault config
// swagger:model InventoryFaultConfig
type InventoryFaultConfig struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Configure controller faults. Field introduced in 20.1.6.
	ControllerFaults *ControllerFaults `json:"controller_faults,omitempty"`

	// Name. Field introduced in 20.1.6.
	Name *string `json:"name,omitempty"`

	// Configure serviceengine faults. Field introduced in 20.1.6.
	ServiceengineFaults *ServiceengineFaults `json:"serviceengine_faults,omitempty"`

	// Tenant. It is a reference to an object of type Tenant. Field introduced in 20.1.6.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// UUID Auto generated. Field introduced in 20.1.6.
	UUID *string `json:"uuid,omitempty"`

	// Configure VirtualService faults. Field introduced in 20.1.6.
	VirtualserviceFaults *VirtualserviceFaults `json:"virtualservice_faults,omitempty"`
}
