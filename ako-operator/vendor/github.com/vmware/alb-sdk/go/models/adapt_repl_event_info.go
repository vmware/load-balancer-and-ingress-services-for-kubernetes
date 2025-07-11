// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AdaptReplEventInfo adapt repl event info
// swagger:model AdaptReplEventInfo
type AdaptReplEventInfo struct {

	// Object config version info. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ObjInfo *ConfigVersionStatus `json:"obj_info,omitempty"`

	// Reason for the replication issues. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// Recommended way to resolve replication issue. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Recommendation *string `json:"recommendation,omitempty"`
}
