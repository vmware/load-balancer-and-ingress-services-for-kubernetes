// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SamlAttribute saml attribute
// swagger:model SamlAttribute
type SamlAttribute struct {

	// SAML Attribute name. Field introduced in 20.1.1.
	AttrName *string `json:"attr_name,omitempty"`

	// SAML Attribute values. Field introduced in 20.1.1.
	AttrValues []string `json:"attr_values,omitempty"`
}
