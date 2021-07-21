// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeGroupOptions se group options
// swagger:model SeGroupOptions
type SeGroupOptions struct {

	// The error recovery action configured for a SE Group. Enum options - ROLLBACK_UPGRADE_OPS_ON_ERROR, SUSPEND_UPGRADE_OPS_ON_ERROR, CONTINUE_UPGRADE_OPS_ON_ERROR. Field introduced in 18.2.6.
	ActionOnError *string `json:"action_on_error,omitempty"`

	// Disable non-disruptive mechanism. Field introduced in 18.2.6.
	Disruptive *bool `json:"disruptive,omitempty"`
}
