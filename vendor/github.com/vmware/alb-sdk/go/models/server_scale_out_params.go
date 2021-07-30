// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServerScaleOutParams server scale out params
// swagger:model ServerScaleOutParams
type ServerScaleOutParams struct {

	// Reason for the manual scaleout.
	Reason *string `json:"reason,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
