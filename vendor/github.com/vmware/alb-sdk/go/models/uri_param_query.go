// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// URIParamQuery URI param query
// swagger:model URIParamQuery
type URIParamQuery struct {

	// Concatenate a *string to the query of the incoming request URI and then use it in the request URI going to the backend server.
	AddString *string `json:"add_string,omitempty"`

	// Use or drop the query of the incoming request URI in the request URI to the backend server.
	KeepQuery *bool `json:"keep_query,omitempty"`
}
