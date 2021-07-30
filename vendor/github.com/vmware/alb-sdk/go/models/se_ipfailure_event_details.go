// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeIpfailureEventDetails se ipfailure event details
// swagger:model SeIpfailureEventDetails
type SeIpfailureEventDetails struct {

	// Mac Address.
	Mac *string `json:"mac,omitempty"`

	// Network UUID.
	NetworkUUID *string `json:"network_uuid,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`

	// Vnic name.
	VnicName *string `json:"vnic_name,omitempty"`
}
