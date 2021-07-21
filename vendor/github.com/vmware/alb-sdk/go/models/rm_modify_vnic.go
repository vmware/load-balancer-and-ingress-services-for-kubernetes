// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RmModifyVnic rm modify vnic
// swagger:model RmModifyVnic
type RmModifyVnic struct {

	// mac_addr of RmModifyVnic.
	MacAddr *string `json:"mac_addr,omitempty"`

	// network_name of RmModifyVnic.
	NetworkName *string `json:"network_name,omitempty"`

	// Unique object identifier of network.
	NetworkUUID *string `json:"network_uuid,omitempty"`
}
