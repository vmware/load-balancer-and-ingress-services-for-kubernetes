// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudConnectorDebugFilter cloud connector debug filter
// swagger:model CloudConnectorDebugFilter
type CloudConnectorDebugFilter struct {

	// filter debugs for an app.
	AppID *string `json:"app_id,omitempty"`

	// Disable SE reboot via cloud connector on HB miss.
	DisableSeReboot *bool `json:"disable_se_reboot,omitempty"`

	// filter debugs for a SE.
	SeID *string `json:"se_id,omitempty"`
}
