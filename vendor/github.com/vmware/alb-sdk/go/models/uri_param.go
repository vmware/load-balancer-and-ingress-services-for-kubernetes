// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// URIParam URI param
// swagger:model URIParam
type URIParam struct {

	// Token config either for the URI components or a constant string. Minimum of 1 items required.
	Tokens []*URIParamToken `json:"tokens,omitempty"`

	// URI param type. Enum options - URI_PARAM_TYPE_TOKENIZED.
	// Required: true
	Type *string `json:"type"`
}
