// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SSLExpireDetails s s l expire details
// swagger:model SSLExpireDetails
type SSLExpireDetails struct {

	// Number of days until certificate is expired. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	DaysLeft *uint32 `json:"days_left,omitempty"`

	// Name of SSL Certificate. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Name *string `json:"name,omitempty"`
}
