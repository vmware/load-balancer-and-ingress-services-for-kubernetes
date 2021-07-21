// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbPoolRuntime gslb pool runtime
// swagger:model GslbPoolRuntime
type GslbPoolRuntime struct {

	// Placeholder for description of property members of obj type GslbPoolRuntime field type str  type object
	Members []*GslbPoolMemberRuntimeInfo `json:"members,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Gslb Pool's consolidated operational status . Field introduced in 18.2.3.
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`
}
