// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RebootData reboot data
// swagger:model RebootData
type RebootData struct {

	// Patch version for which reboot flag need to be computed. Field introduced in 18.2.8, 20.1.1.
	PatchVersion *string `json:"patch_version,omitempty"`

	// This variable tells whether reboot has to be performed. Field introduced in 18.2.8, 20.1.1.
	Reboot *bool `json:"reboot,omitempty"`
}
