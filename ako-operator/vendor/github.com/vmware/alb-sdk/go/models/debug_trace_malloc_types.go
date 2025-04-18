// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// DebugTraceMallocTypes debug trace malloc types
// swagger:model DebugTraceMallocTypes
type DebugTraceMallocTypes struct {

	// Memory type to be traced for se_malloc and se_free. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MallocTypeIndex *uint32 `json:"malloc_type_index,omitempty"`
}
