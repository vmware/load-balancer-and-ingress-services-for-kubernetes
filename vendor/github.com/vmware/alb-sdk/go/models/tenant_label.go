// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// TenantLabel tenant label
// swagger:model TenantLabel
type TenantLabel struct {

	// Label key string. Field introduced in 20.1.2.
	// Required: true
	Key *string `json:"key"`

	// Label value string. Field introduced in 20.1.2.
	Value *string `json:"value,omitempty"`
}
