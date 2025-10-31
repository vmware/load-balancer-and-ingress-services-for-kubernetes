// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// URIParamToken URI param token
// swagger:model URIParamToken
type URIParamToken struct {

	// Index of the ending token in the incoming URI. Allowed values are 0-65534. Special values are 65535 - end of string. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EndIndex *uint32 `json:"end_index,omitempty"`

	// Index of the starting token in the incoming URI. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StartIndex *uint32 `json:"start_index,omitempty"`

	// Constant *string to use as a token. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StrValue *string `json:"str_value,omitempty"`

	// Token type for constructing the URI. Enum options - URI_TOKEN_TYPE_HOST, URI_TOKEN_TYPE_PATH, URI_TOKEN_TYPE_STRING, URI_TOKEN_TYPE_STRING_GROUP, URI_TOKEN_TYPE_REGEX, URI_TOKEN_TYPE_REGEX_QUERY. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`
}
