// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthMatchAttribute auth match attribute
// swagger:model AuthMatchAttribute
type AuthMatchAttribute struct {

	// rule match criteria. Enum options - AUTH_MATCH_CONTAINS, AUTH_MATCH_DOES_NOT_CONTAIN, AUTH_MATCH_REGEX.
	// Required: true
	Criteria *string `json:"criteria"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// values of AuthMatchAttribute.
	Values []string `json:"values,omitempty"`
}
