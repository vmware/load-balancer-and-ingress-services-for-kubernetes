// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IcapNsxDefenderConfig icap nsx defender config
// swagger:model IcapNsxDefenderConfig
type IcapNsxDefenderConfig struct {

	// URL to get details from NSXDefender using task_uuid for a particular request. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StatusURL *string `json:"status_url,omitempty"`
}
