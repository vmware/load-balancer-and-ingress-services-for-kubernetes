// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FailActionHTTPRedirect fail action HTTP redirect
// swagger:model FailActionHTTPRedirect
type FailActionHTTPRedirect struct {

	// host of FailActionHTTPRedirect.
	// Required: true
	Host *string `json:"host"`

	// path of FailActionHTTPRedirect.
	Path *string `json:"path,omitempty"`

	//  Enum options - HTTP, HTTPS. Allowed in Basic(Allowed values- HTTP) edition, Enterprise edition. Special default for Basic edition is HTTP, Enterprise is HTTPS.
	Protocol *string `json:"protocol,omitempty"`

	// query of FailActionHTTPRedirect.
	Query *string `json:"query,omitempty"`

	//  Enum options - HTTP_REDIRECT_STATUS_CODE_301, HTTP_REDIRECT_STATUS_CODE_302, HTTP_REDIRECT_STATUS_CODE_307. Allowed in Basic(Allowed values- HTTP_REDIRECT_STATUS_CODE_302) edition, Enterprise edition.
	StatusCode *string `json:"status_code,omitempty"`
}
