// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// NsxtSetup nsxt setup
// swagger:model NsxtSetup
type NsxtSetup struct {

	// cc_id of NsxtSetup.
	CcID *string `json:"cc_id,omitempty"`

	// reason of NsxtSetup.
	Reason *string `json:"reason,omitempty"`

	// status of NsxtSetup.
	Status *string `json:"status,omitempty"`

	// transportzone_id of NsxtSetup.
	TransportzoneID *string `json:"transportzone_id,omitempty"`
}
