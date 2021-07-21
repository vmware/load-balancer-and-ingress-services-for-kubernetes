// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsSeVnic vs se vnic
// swagger:model VsSeVnic
type VsSeVnic struct {

	// lif of VsSeVnic.
	Lif *string `json:"lif,omitempty"`

	// mac of VsSeVnic.
	// Required: true
	Mac *string `json:"mac"`

	//  Enum options - VNIC_TYPE_FE, VNIC_TYPE_BE.
	// Required: true
	Type *string `json:"type"`
}
