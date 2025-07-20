// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FalsePositiveResultHeader false positive result header
// swagger:model FalsePositiveResultHeader
type FalsePositiveResultHeader struct {

	// Time that Analytics Engine ends to analytics for this false positive result. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndAnalysisTime *string `json:"end_analysis_time,omitempty"`

	// First received data time that Analytics Engine uses to analysis for this false positive result. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	FirstDataReceivedTime *string `json:"first_data_received_time,omitempty"`

	// Last received data time that Analytics Engine uses to analysis for this false positive result. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LastDataReceivedTime *string `json:"last_data_received_time,omitempty"`

	// Time that Analytics Engine starts to analytics for this false positive result. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StartAnalysisTime *string `json:"start_analysis_time,omitempty"`

	// Total data amount Analytics Engine uses to analytics for this false positive result. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TransactionsCount *int64 `json:"transactions_count,omitempty"`
}
