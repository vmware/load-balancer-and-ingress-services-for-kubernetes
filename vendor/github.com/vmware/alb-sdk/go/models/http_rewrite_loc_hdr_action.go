// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPRewriteLocHdrAction HTTP rewrite loc hdr action
// swagger:model HTTPRewriteLocHdrAction
type HTTPRewriteLocHdrAction struct {

	// Host config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Host *URIParam `json:"host,omitempty"`

	// Keep or drop the query from the server side redirect URI. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeepQuery *bool `json:"keep_query,omitempty"`

	// Path config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *URIParam `json:"path,omitempty"`

	// Port to use in the redirected URI. Allowed values are 1-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port *uint32 `json:"port,omitempty"`

	// HTTP protocol type. Enum options - HTTP, HTTPS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Protocol *string `json:"protocol"`
}
