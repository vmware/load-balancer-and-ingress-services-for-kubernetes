// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HTTPRedirectAction HTTP redirect action
// swagger:model HTTPRedirectAction
type HTTPRedirectAction struct {

	// Add a query *string to the redirect URI. If keep_query is set, concatenates the add_string to the query of the incoming request. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AddString *string `json:"add_string,omitempty"`

	// Host config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Host *URIParam `json:"host,omitempty"`

	// Keep or drop the query of the incoming request URI in the redirected URI. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	KeepQuery *bool `json:"keep_query,omitempty"`

	// Path config. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *URIParam `json:"path,omitempty"`

	// Port to which redirect the request. Allowed values are 1-65535. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Port uint32 `json:"port,omitempty"`

	// Protocol type. Enum options - HTTP, HTTPS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Protocol *string `json:"protocol"`

	// HTTP redirect status code. Enum options - HTTP_REDIRECT_STATUS_CODE_301, HTTP_REDIRECT_STATUS_CODE_302, HTTP_REDIRECT_STATUS_CODE_307. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StatusCode *string `json:"status_code,omitempty"`
}
