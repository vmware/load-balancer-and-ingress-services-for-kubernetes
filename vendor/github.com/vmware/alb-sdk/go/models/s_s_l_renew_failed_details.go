// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLRenewFailedDetails s s l renew failed details
// swagger:model SSLRenewFailedDetails
type SSLRenewFailedDetails struct {

	// Error when renewing certificate.
	Error *string `json:"error,omitempty"`

	// Name of SSL Certificate.
	Name *string `json:"name,omitempty"`
}
