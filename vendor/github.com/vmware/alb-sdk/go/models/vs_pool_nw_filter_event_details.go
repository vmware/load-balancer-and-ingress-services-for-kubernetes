// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsPoolNwFilterEventDetails vs pool nw filter event details
// swagger:model VsPoolNwFilterEventDetails
type VsPoolNwFilterEventDetails struct {

	// filter of VsPoolNwFilterEventDetails.
	// Required: true
	Filter *string `json:"filter"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// network of VsPoolNwFilterEventDetails.
	// Required: true
	Network *string `json:"network"`
}
