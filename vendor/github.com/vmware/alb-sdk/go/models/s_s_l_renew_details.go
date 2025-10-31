// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLRenewDetails s s l renew details
// swagger:model SSLRenewDetails
type SSLRenewDetails struct {

	// Name of SSL Certificate. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`
}
