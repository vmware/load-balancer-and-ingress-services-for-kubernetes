// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeDupipEventDetails se dupip event details
// swagger:model SeDupipEventDetails
type SeDupipEventDetails struct {

	// Mac Address.
	LocalMac *string `json:"local_mac,omitempty"`

	// Mac Address.
	RemoteMac *string `json:"remote_mac,omitempty"`

	// Vnic IP.
	VnicIP *string `json:"vnic_ip,omitempty"`

	// Vnic name.
	VnicName *string `json:"vnic_name,omitempty"`
}
