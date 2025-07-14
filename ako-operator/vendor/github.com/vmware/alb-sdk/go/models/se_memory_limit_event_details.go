// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeMemoryLimitEventDetails se memory limit event details
// swagger:model SeMemoryLimitEventDetails
type SeMemoryLimitEventDetails struct {

	// Current status of config memory. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfigMemoryStatus *string `json:"config_memory_status,omitempty"`

	// Heap config memory hard limit. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeapConfigHardLimit *uint32 `json:"heap_config_hard_limit,omitempty"`

	// Heap config memory soft limit. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeapConfigSoftLimit *uint32 `json:"heap_config_soft_limit,omitempty"`

	// Config memory usage in heap memory. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeapConfigUsage *uint32 `json:"heap_config_usage,omitempty"`

	// Connection memory usage in heap memory. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	HeapConnUsage *uint32 `json:"heap_conn_usage,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeRef *string `json:"se_ref,omitempty"`

	// Current shm config memory hard limit. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ShmConfigHardLimit *uint32 `json:"shm_config_hard_limit,omitempty"`

	// Current shm config memory soft limit. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ShmConfigSoftLimit *uint32 `json:"shm_config_soft_limit,omitempty"`

	// Config memory usage in shared memory. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ShmConfigUsage *uint32 `json:"shm_config_usage,omitempty"`

	// Connection memory usage in shared memory. Field introduced in 18.2.2. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ShmConnUsage *uint32 `json:"shm_conn_usage,omitempty"`
}
