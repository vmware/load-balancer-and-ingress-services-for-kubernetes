// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VIGuestvNicIPAddr v i guestv nic IP addr
// swagger:model VIGuestvNicIPAddr
type VIGuestvNicIPAddr struct {

	// ip_addr of VIGuestvNicIPAddr.
	// Required: true
	IPAddr *string `json:"ip_addr"`

	// Number of mask.
	// Required: true
	Mask *int32 `json:"mask"`
}
