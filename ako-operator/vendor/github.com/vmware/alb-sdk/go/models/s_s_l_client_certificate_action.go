// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLClientCertificateAction s s l client certificate action
// swagger:model SSLClientCertificateAction
type SSLClientCertificateAction struct {

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CloseConnection *bool `json:"close_connection,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Headers []*SSLClientRequestHeader `json:"headers,omitempty"`
}
