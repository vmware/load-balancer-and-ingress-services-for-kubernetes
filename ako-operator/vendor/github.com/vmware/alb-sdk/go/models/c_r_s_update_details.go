// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CRSUpdateDetails c r s update details
// swagger:model CRSUpdateDetails
type CRSUpdateDetails struct {

	// List of all available CRS updates. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CrsInfo []*CRSDetails `json:"crs_info,omitempty"`
}
