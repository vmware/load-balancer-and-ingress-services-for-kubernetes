// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SystemUpgradeDetails system upgrade details
// swagger:model SystemUpgradeDetails
type SystemUpgradeDetails struct {

	// Upgrade status.
	UpgradeStatus *SystemUpgradeState `json:"upgrade_status,omitempty"`
}
