// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// URIInfo URI info
// swagger:model URIInfo
type URIInfo struct {

	// Information about various params under a URI. Field introduced in 20.1.1.
	ParamInfo []*ParamInfo `json:"param_info,omitempty"`

	// Total number of URI hits. Field introduced in 20.1.1.
	URIHits *int64 `json:"uri_hits,omitempty"`

	// URI name. Field introduced in 20.1.1.
	URIKey *string `json:"uri_key,omitempty"`
}
