// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// HttpsecurityAction httpsecurity action
// swagger:model HTTPSecurityAction
type HttpsecurityAction struct {

	// Type of the security action to perform. Enum options - HTTP_SECURITY_ACTION_CLOSE_CONN, HTTP_SECURITY_ACTION_SEND_RESPONSE, HTTP_SECURITY_ACTION_ALLOW, HTTP_SECURITY_ACTION_REDIRECT_TO_HTTPS, HTTP_SECURITY_ACTION_RATE_LIMIT, HTTP_SECURITY_ACTION_REQUEST_CHECK_ICAP. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- HTTP_SECURITY_ACTION_CLOSE_CONN,HTTP_SECURITY_ACTION_SEND_RESPONSE,HTTP_SECURITY_ACTION_REDIRECT_TO_HTTPS), Basic edition(Allowed values- HTTP_SECURITY_ACTION_CLOSE_CONN,HTTP_SECURITY_ACTION_SEND_RESPONSE,HTTP_SECURITY_ACTION_REDIRECT_TO_HTTPS), Enterprise with Cloud Services edition.
	// Required: true
	Action *string `json:"action"`

	// File to be used for generating HTTP local response. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	File *HTTPLocalFile `json:"file,omitempty"`

	// Secure SSL/TLS port to redirect the HTTP request to. Allowed values are 1-65535. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HTTPSPort uint32 `json:"https_port,omitempty"`

	// Rate limiting configuration for this action. Field introduced in 18.2.9. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RateProfile *HttpsecurityActionRateProfile `json:"rate_profile,omitempty"`

	// HTTP status code to use for local response. Enum options - HTTP_LOCAL_RESPONSE_STATUS_CODE_200, HTTP_LOCAL_RESPONSE_STATUS_CODE_204, HTTP_LOCAL_RESPONSE_STATUS_CODE_403, HTTP_LOCAL_RESPONSE_STATUS_CODE_404, HTTP_LOCAL_RESPONSE_STATUS_CODE_429, HTTP_LOCAL_RESPONSE_STATUS_CODE_501. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	StatusCode *string `json:"status_code,omitempty"`
}
