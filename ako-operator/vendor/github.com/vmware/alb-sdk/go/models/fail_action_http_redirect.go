// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FailActionHTTPRedirect fail action HTTP redirect
// swagger:model FailActionHTTPRedirect
type FailActionHTTPRedirect struct {

	// The host to which the redirect request is sent. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Host *string `json:"host"`

	// Path configuration for the redirect request. If not set the path from the original request's URI is preserved in the redirect on pool failure. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Path *string `json:"path,omitempty"`

	//  Enum options - HTTP, HTTPS. Allowed in Enterprise edition with any value, Basic edition(Allowed values- HTTP), Essentials, Enterprise with Cloud Services edition. Special default for Basic edition is HTTP, Enterprise is HTTPS.
	Protocol *string `json:"protocol,omitempty"`

	// Query configuration for the redirect request URI. If not set, the query from the original request's URI is preserved in the redirect on pool failure. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Query *string `json:"query,omitempty"`

	//  Enum options - HTTP_REDIRECT_STATUS_CODE_301, HTTP_REDIRECT_STATUS_CODE_302, HTTP_REDIRECT_STATUS_CODE_307. Allowed in Enterprise edition with any value, Basic edition(Allowed values- HTTP_REDIRECT_STATUS_CODE_302), Essentials, Enterprise with Cloud Services edition.
	StatusCode *string `json:"status_code,omitempty"`
}
