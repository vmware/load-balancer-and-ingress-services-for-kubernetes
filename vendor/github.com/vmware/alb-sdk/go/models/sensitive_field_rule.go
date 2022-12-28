// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SensitiveFieldRule sensitive field rule
// swagger:model SensitiveFieldRule
type SensitiveFieldRule struct {

	// Action for the matched log field, for instance the matched field can be removed or masked off. Enum options - LOG_FIELD_REMOVE, LOG_FIELD_MASKOFF. Field introduced in 17.2.10, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Action *string `json:"action,omitempty"`

	// Enable rule to match the sensitive fields. Field introduced in 17.2.10, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enabled *bool `json:"enabled,omitempty"`

	// Index of the rule. Field introduced in 17.2.10, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Index *int32 `json:"index,omitempty"`

	// Criterion to use for matching in the Log. Field introduced in 17.2.10, 18.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Match *StringMatch `json:"match,omitempty"`

	// Name of the rule. Field introduced in 17.2.10, 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`
}
