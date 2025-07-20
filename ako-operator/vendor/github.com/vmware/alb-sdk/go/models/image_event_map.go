// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ImageEventMap image event map
// swagger:model ImageEventMap
type ImageEventMap struct {

	// List of all events node wise. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	NodesEvents []*ImageEvent `json:"nodes_events,omitempty"`

	// List of all events node wise. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SubEvents []*ImageEvent `json:"sub_events,omitempty"`

	// Name representing the task. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	TaskName *string `json:"task_name,omitempty"`
}
