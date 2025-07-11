// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VHMatchRule v h match rule
// swagger:model VHMatchRule
type VHMatchRule struct {

	// Match criteria for selecting the virtual service. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Matches *MatchTarget `json:"matches"`

	// Name for the match rule. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Name *string `json:"name"`
}
