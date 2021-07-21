// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLClientCertificateAction s s l client certificate action
// swagger:model SSLClientCertificateAction
type SSLClientCertificateAction struct {

	// Placeholder for description of property close_connection of obj type SSLClientCertificateAction field type str  type boolean
	CloseConnection *bool `json:"close_connection,omitempty"`

	// Placeholder for description of property headers of obj type SSLClientCertificateAction field type str  type object
	Headers []*SSLClientRequestHeader `json:"headers,omitempty"`
}
