// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPRequestPolicy HTTP request policy
// swagger:model HTTPRequestPolicy
type HTTPRequestPolicy struct {

	// Add rules to the HTTP request policy. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Rules []*HTTPRequestRule `json:"rules,omitempty"`
}
