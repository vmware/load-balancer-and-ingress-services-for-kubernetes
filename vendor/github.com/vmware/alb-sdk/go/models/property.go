// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// Property property
// swagger:model Property
type Property struct {

	// Property name. Field introduced in 17.2.1.
	// Required: true
	Name *string `json:"name"`

	// Property value. Field introduced in 17.2.1.
	Value *string `json:"value,omitempty"`
}
