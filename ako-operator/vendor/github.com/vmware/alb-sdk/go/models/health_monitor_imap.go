// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HealthMonitorImap health monitor imap
// swagger:model HealthMonitorImap
type HealthMonitorImap struct {

	// Folder to access. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Folder *string `json:"folder,omitempty"`

	// SSL attributes for IMAPS monitor. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SslAttributes *HealthMonitorSSlattributes `json:"ssl_attributes,omitempty"`
}
