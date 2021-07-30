// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLRevokedDetails s s l revoked details
// swagger:model SSLRevokedDetails
type SSLRevokedDetails struct {

	// Name of SSL Certificate. Field introduced in 20.1.1.
	Name *string `json:"name,omitempty"`

	// Certificate revocation reason provided by OCSP Responder. Field introduced in 20.1.1.
	Reason *string `json:"reason,omitempty"`
}
