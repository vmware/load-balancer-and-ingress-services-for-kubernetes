// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VsEvStatus vs ev status
// swagger:model VsEvStatus
type VsEvStatus struct {

	// notes of VsEvStatus.
	Notes []string `json:"notes,omitempty"`

	// request of VsEvStatus.
	Request *string `json:"request,omitempty"`

	// result of VsEvStatus.
	Result *string `json:"result,omitempty"`
}
