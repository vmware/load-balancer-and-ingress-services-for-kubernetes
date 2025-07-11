// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// L4ConnectionPolicy l4 connection policy
// swagger:model L4ConnectionPolicy
type L4ConnectionPolicy struct {

	// Rules to apply when a new transport connection is setup. Field introduced in 17.2.7. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Rules []*L4Rule `json:"rules,omitempty"`
}
