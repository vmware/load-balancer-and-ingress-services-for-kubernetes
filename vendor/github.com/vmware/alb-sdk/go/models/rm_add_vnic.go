// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RmAddVnic rm add vnic
// swagger:model RmAddVnic
type RmAddVnic struct {

	// mac_addr associated with the network. Field introduced in 21.1.1.
	MacAddr *string `json:"mac_addr,omitempty"`

	// network_name of RmAddVnic.
	NetworkName *string `json:"network_name,omitempty"`

	// Unique object identifier of network.
	NetworkUUID *string `json:"network_uuid,omitempty"`

	// subnet of RmAddVnic.
	Subnet *string `json:"subnet,omitempty"`
}
