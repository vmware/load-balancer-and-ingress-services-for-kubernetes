// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// OmittedWafLogStats omitted waf log stats
// swagger:model OmittedWafLogStats
type OmittedWafLogStats struct {

	// The total count of omitted match element logs in all rules. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchElements *uint32 `json:"match_elements,omitempty"`

	// The total count of omitted rule logs. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Rules *uint32 `json:"rules,omitempty"`
}
