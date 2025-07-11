// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CRSDeploymentSuccess c r s deployment success
// swagger:model CRSDeploymentSuccess
type CRSDeploymentSuccess struct {

	// List of all installed CRS updates. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CrsInfo []*CRSDetails `json:"crs_info,omitempty"`
}
