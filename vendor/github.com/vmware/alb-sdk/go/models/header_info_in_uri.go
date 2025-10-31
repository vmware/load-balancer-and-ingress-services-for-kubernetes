// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HeaderInfoInURI header info in URI
// swagger:model HeaderInfoInURI
type HeaderInfoInURI struct {

	// Header field name in hitted signature rule match_element. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HeaderFieldName *string `json:"header_field_name,omitempty"`

	// Header field value in hitted signature rule match_element. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Value *string `json:"value,omitempty"`
}
