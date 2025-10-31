// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeGroupResumeOptions se group resume options
// swagger:model SeGroupResumeOptions
type SeGroupResumeOptions struct {

	// The error recovery action configured for a SE Group. Enum options - ROLLBACK_UPGRADE_OPS_ON_ERROR, SUSPEND_UPGRADE_OPS_ON_ERROR, CONTINUE_UPGRADE_OPS_ON_ERROR. Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ActionOnError *string `json:"action_on_error,omitempty"`

	// Allow disruptive mechanism. Field introduced in 18.2.8, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Disruptive *bool `json:"disruptive,omitempty"`

	// Skip upgrade on suspended SE(s). Field introduced in 18.2.6. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SkipSuspended *bool `json:"skip_suspended,omitempty"`
}
