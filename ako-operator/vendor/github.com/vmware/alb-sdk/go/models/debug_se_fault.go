// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugSeFault debug se fault
// swagger:model DebugSeFault
type DebugSeFault struct {

	// Set of faults to enable/disable. Field introduced in 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Faults []*SeFault `json:"faults,omitempty"`

	// Fail SE malloc type at this frequency. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMallocFailFrequency *uint32 `json:"se_malloc_fail_frequency,omitempty"`

	// Fail this SE malloc type. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMallocFailType *uint32 `json:"se_malloc_fail_type,omitempty"`

	// Toggle assert on mbuf cluster sanity check fail. Field introduced in 17.2.13,18.1.3,18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeMbufClSanity *bool `json:"se_mbuf_cl_sanity,omitempty"`

	// Fail SE SHM malloc type at this frequency. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeShmMallocFailFrequency *uint32 `json:"se_shm_malloc_fail_frequency,omitempty"`

	// Fail this SE SHM malloc type. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeShmMallocFailType *uint32 `json:"se_shm_malloc_fail_type,omitempty"`

	// Fail SE WAF allocation at this frequency. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeWafAllocFailFrequency *uint32 `json:"se_waf_alloc_fail_frequency,omitempty"`

	// Fail SE WAF learning allocation at this frequency. Field introduced in 18.1.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeWafLearningAllocFailFrequency *uint32 `json:"se_waf_learning_alloc_fail_frequency,omitempty"`
}
