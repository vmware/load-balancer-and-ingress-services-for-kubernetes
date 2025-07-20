// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ErrorPage error page
// swagger:model ErrorPage
type ErrorPage struct {

	// Enable or disable the error page. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Enable *bool `json:"enable,omitempty"`

	// Custom error page body used to sent to the client. It is a reference to an object of type ErrorPageBody. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ErrorPageBodyRef *string `json:"error_page_body_ref,omitempty"`

	// Redirect sent to client when match. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ErrorRedirect *string `json:"error_redirect,omitempty"`

	// Index of the error page. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Index *int32 `json:"index,omitempty"`

	// Add match criteria for http status codes to the error page. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Match *HttpstatusMatch `json:"match,omitempty"`
}
