// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// URIInfo URI info
// swagger:model URIInfo
type URIInfo struct {

	// Information about various params under a URI. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ParamInfo []*ParamInfo `json:"param_info,omitempty"`

	// Total number of URI hits. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URIHits *uint64 `json:"uri_hits,omitempty"`

	// URI name. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	URIKey *string `json:"uri_key,omitempty"`
}
