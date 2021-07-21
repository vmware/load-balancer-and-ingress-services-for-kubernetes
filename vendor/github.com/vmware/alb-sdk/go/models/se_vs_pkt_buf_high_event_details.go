// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeVsPktBufHighEventDetails se vs pkt buf high event details
// swagger:model SeVsPktBufHighEventDetails
type SeVsPktBufHighEventDetails struct {

	// Current packet buffer usage of the VS.
	CurrentValue *int32 `json:"current_value,omitempty"`

	// Buffer usage threshold value.
	Threshold *int32 `json:"threshold,omitempty"`

	// Virtual Service name. It is a reference to an object of type VirtualService.
	VirtualService *string `json:"virtual_service,omitempty"`
}
