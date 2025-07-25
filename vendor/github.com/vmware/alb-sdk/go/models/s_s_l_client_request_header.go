// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLClientRequestHeader s s l client request header
// swagger:model SSLClientRequestHeader
type SSLClientRequestHeader struct {

	// If this header exists, reset the connection. If the ssl variable is specified, add a header with this value. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestHeader *string `json:"request_header,omitempty"`

	// Set the request header with the value as indicated by this SSL variable. Eg. send the whole certificate in PEM format. Enum options - HTTP_POLICY_VAR_CLIENT_IP, HTTP_POLICY_VAR_VS_PORT, HTTP_POLICY_VAR_VS_IP, HTTP_POLICY_VAR_HTTP_HDR, HTTP_POLICY_VAR_SSL_CLIENT_FINGERPRINT, HTTP_POLICY_VAR_SSL_CLIENT_SERIAL, HTTP_POLICY_VAR_SSL_CLIENT_ISSUER, HTTP_POLICY_VAR_SSL_CLIENT_SUBJECT, HTTP_POLICY_VAR_SSL_CLIENT_RAW, HTTP_POLICY_VAR_SSL_PROTOCOL, HTTP_POLICY_VAR_SSL_SERVER_NAME, HTTP_POLICY_VAR_USER_NAME, HTTP_POLICY_VAR_SSL_CIPHER, HTTP_POLICY_VAR_REQUEST_ID, HTTP_POLICY_VAR_SSL_CLIENT_VERSION, HTTP_POLICY_VAR_SSL_CLIENT_SIGALG, HTTP_POLICY_VAR_SSL_CLIENT_NOTVALIDBEFORE, HTTP_POLICY_VAR_SSL_CLIENT_NOTVALIDAFTER, HTTP_POLICY_VAR_SSL_CLIENT_ESCAPED, HTTP_POLICY_VAR_SOURCE_IP. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RequestHeaderValue *string `json:"request_header_value,omitempty"`
}
