// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPRewriteURLAction HTTP rewrite URL action
// swagger:model HTTPRewriteURLAction
type HTTPRewriteURLAction struct {

	// Host config.
	HostHdr *URIParam `json:"host_hdr,omitempty"`

	// Path config.
	Path *URIParam `json:"path,omitempty"`

	// Query config.
	Query *URIParamQuery `json:"query,omitempty"`
}
