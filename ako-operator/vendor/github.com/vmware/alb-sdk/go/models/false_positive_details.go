// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FalsePositiveDetails false positive details
// swagger:model FalsePositiveDetails
type FalsePositiveDetails struct {

	// False Positive result details. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FalsePositiveResults []*FalsePositiveResult `json:"false_positive_results,omitempty"`

	// VirtualService Name for which False Positive is being generated. Field introduced in 30.1.2. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsName *string `json:"vs_name,omitempty"`

	// VirtualService UUID for which False Positive is being generated. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	VsUUID *string `json:"vs_uuid,omitempty"`
}
