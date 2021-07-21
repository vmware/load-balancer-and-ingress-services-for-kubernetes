// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ServerScaleInParams server scale in params
// swagger:model ServerScaleInParams
type ServerScaleInParams struct {

	// Reason for the manual scalein.
	Reason *string `json:"reason,omitempty"`

	// List of server IDs that should be scaled in.
	Servers []*ServerID `json:"servers,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
