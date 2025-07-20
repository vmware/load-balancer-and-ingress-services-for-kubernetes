// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudAutoscalingConfigFailureDetails cloud autoscaling config failure details
// swagger:model CloudAutoscalingConfigFailureDetails
type CloudAutoscalingConfigFailureDetails struct {

	// Cloud UUID. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	CcID *string `json:"cc_id,omitempty"`

	// Failure reason if Autoscaling configuration fails. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ErrorString *string `json:"error_string,omitempty"`
}
