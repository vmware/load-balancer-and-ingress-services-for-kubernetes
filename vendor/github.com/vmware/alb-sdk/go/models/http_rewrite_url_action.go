// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPRewriteURLAction HTTP rewrite URL action
// swagger:model HTTPRewriteURLAction
type HTTPRewriteURLAction struct {

	// Host config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HostHdr *URIParam `json:"host_hdr,omitempty"`

	// Path config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *URIParam `json:"path,omitempty"`

	// Query config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Query *URIParamQuery `json:"query,omitempty"`
}
