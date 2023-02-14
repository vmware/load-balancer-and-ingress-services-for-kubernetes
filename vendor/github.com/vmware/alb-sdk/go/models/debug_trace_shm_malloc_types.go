// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugTraceShmMallocTypes debug trace shm malloc types
// swagger:model DebugTraceShmMallocTypes
type DebugTraceShmMallocTypes struct {

	// Memory type to be traced for se_shmalloc and se_shmfree. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ShmMallocTypeIndex *int32 `json:"shm_malloc_type_index,omitempty"`
}
