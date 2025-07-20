// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ReportSummary report summary
// swagger:model ReportSummary
type ReportSummary struct {

	// Detailed description of the report. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Preview of the operations performed in the report. Ex  Upgrade Pre-check Previews. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Previews []string `json:"previews,omitempty"`

	// User friendly title for the report. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Title *string `json:"title,omitempty"`
}
