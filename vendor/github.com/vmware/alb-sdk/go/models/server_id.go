// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServerID server Id
// swagger:model ServerId
type ServerID struct {

	// This is the external cloud uuid of the Pool server.
	ExternalUUID *string `json:"external_uuid,omitempty"`

	// Placeholder for description of property ip of obj type ServerId field type str  type object
	// Required: true
	IP *IPAddr `json:"ip"`

	// Number of port.
	// Required: true
	Port *int32 `json:"port"`
}
