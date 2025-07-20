// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TenantLabel tenant label
// swagger:model TenantLabel
type TenantLabel struct {

	// Label key string. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Key *string `json:"key"`

	// Label value string. Field introduced in 20.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}
