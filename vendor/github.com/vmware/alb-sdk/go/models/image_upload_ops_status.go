// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ImageUploadOpsStatus image upload ops status
// swagger:model ImageUploadOpsStatus
type ImageUploadOpsStatus struct {

	// The last time the state changed. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	LastChangedTime *TimeStamp `json:"last_changed_time,omitempty"`

	// Descriptive reason for the state of the image. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Reason *string `json:"reason,omitempty"`

	// Current fsm-state of image upload operation. Enum options - IMAGE_FSM_STARTED, IMAGE_FSM_IN_PROGRESS, IMAGE_FSM_COMPLETED, IMAGE_FSM_FAILED. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`
}
