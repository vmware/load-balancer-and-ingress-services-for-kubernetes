// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RateLimiterProperties rate limiter properties
// swagger:model RateLimiterProperties
type RateLimiterProperties struct {

	// Number of stages in msf rate limiter. Allowed values are 1-2. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MsfNumStages *uint32 `json:"msf_num_stages,omitempty"`

	// Each stage size in msf rate limiter. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MsfStageSize *uint64 `json:"msf_stage_size,omitempty"`
}
