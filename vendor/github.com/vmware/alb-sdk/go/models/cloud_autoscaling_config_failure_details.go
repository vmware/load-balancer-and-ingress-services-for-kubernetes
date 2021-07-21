// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CloudAutoscalingConfigFailureDetails cloud autoscaling config failure details
// swagger:model CloudAutoscalingConfigFailureDetails
type CloudAutoscalingConfigFailureDetails struct {

	// Cloud UUID. Field introduced in 20.1.1.
	CcID *string `json:"cc_id,omitempty"`

	// Failure reason if Autoscaling configuration fails. Field introduced in 20.1.1.
	ErrorString *string `json:"error_string,omitempty"`
}
