// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPResponsePolicy HTTP response policy
// swagger:model HTTPResponsePolicy
type HTTPResponsePolicy struct {

	// Add rules to the HTTP response policy. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Rules []*HTTPResponseRule `json:"rules,omitempty"`
}
