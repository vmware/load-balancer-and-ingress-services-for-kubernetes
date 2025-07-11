// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugTraceMemory debug trace memory
// swagger:model DebugTraceMemory
type DebugTraceMemory struct {

	// Memory type to be traced for se_malloc and se_free. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TraceMallocTypes []*DebugTraceMallocTypes `json:"trace_malloc_types,omitempty"`

	// Memory type to be traced for se_shm_malloc and se_shm_free. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TraceShmMallocTypes []*DebugTraceShmMallocTypes `json:"trace_shm_malloc_types,omitempty"`
}
