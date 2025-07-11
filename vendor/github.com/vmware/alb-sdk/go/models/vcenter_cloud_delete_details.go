// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// VcenterCloudDeleteDetails vcenter cloud delete details
// swagger:model VcenterCloudDeleteDetails
type VcenterCloudDeleteDetails struct {

	// Cloud id. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CcID *string `json:"cc_id,omitempty"`

	// Objects having reference to the cloud. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Objects *string `json:"objects,omitempty"`
}
