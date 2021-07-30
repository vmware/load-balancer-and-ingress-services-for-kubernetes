// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorImap health monitor imap
// swagger:model HealthMonitorImap
type HealthMonitorImap struct {

	// Folder to access. Field introduced in 20.1.5.
	Folder *string `json:"folder,omitempty"`

	// SSL attributes for IMAPS monitor. Field introduced in 20.1.5.
	SslAttributes *HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
}
