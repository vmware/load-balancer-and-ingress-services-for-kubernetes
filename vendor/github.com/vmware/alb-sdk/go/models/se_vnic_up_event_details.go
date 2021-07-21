// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeVnicUpEventDetails se vnic up event details
// swagger:model SeVnicUpEventDetails
type SeVnicUpEventDetails struct {

	// Vnic name.
	IfName *string `json:"if_name,omitempty"`

	// Vnic linux name.
	LinuxName *string `json:"linux_name,omitempty"`

	// Mac Address.
	Mac *string `json:"mac,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`
}
