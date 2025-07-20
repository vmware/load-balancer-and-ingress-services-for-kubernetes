// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudConnectorDebugFilter cloud connector debug filter
// swagger:model CloudConnectorDebugFilter
type CloudConnectorDebugFilter struct {

	// filter debugs for an app. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AppID *string `json:"app_id,omitempty"`

	// Disable SE reboot via cloud connector on HB miss. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DisableSeReboot *bool `json:"disable_se_reboot,omitempty"`

	// filter debugs for a SE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeID *string `json:"se_id,omitempty"`
}
