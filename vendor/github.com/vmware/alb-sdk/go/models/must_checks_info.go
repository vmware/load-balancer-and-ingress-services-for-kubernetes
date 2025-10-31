// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MustChecksInfo must checks info
// swagger:model MustChecksInfo
type MustChecksInfo struct {

	// Syserr status code of the must check. Enum options - SYSERR_SUCCESS, SYSERR_FAILURE, SYSERR_OUT_OF_MEMORY, SYSERR_NO_ENT, SYSERR_INVAL, SYSERR_ACCESS, SYSERR_FAULT, SYSERR_IO, SYSERR_TIMEOUT, SYSERR_NOT_SUPPORTED, SYSERR_NOT_READY, SYSERR_UPGRADE_IN_PROGRESS, SYSERR_WARM_START_IN_PROGRESS, SYSERR_TRY_AGAIN, SYSERR_NOT_UPGRADING, SYSERR_PENDING, SYSERR_EVENT_GEN_FAILURE, SYSERR_CONFIG_PARAM_MISSING, SYSERR_RANGE, SYSERR_BAD_REQUEST.... Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CheckCode *string `json:"check_code,omitempty"`

	// Reason for must check failure. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Description *string `json:"description,omitempty"`

	// Additional details of the must check. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Details []string `json:"details,omitempty"`

	// Time taken to complete must check in seconds. Field introduced in 22.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Duration *uint32 `json:"duration,omitempty"`

	// Time at which execution of must check was completed. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndTime *string `json:"end_time,omitempty"`

	// Error/Failure details of the must check. Field introduced in 22.1.6. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ErrorDetails []string `json:"error_details,omitempty"`

	// Time at which execution of must check was started. Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StartTime *string `json:"start_time,omitempty"`

	// The must check operations current fsm-state. Enum options - UPGRADE_FSM_INIT, UPGRADE_FSM_STARTED, UPGRADE_FSM_WAITING, UPGRADE_FSM_IN_PROGRESS, UPGRADE_FSM_ENQUEUED, UPGRADE_FSM_ERROR, UPGRADE_FSM_SUSPENDED, UPGRADE_FSM_ENQUEUE_FAILED, UPGRADE_FSM_PAUSED, UPGRADE_FSM_COMPLETED, UPGRADE_FSM_ABORT_IN_PROGRESS, UPGRADE_FSM_ABORTED, UPGRADE_FSM_SE_UPGRADE_IN_PROGRESS, UPGRADE_FSM_CONTROLLER_COMPLETED, UPGRADE_FSM_DUMMY_3, UPGRADE_FSM_DUMMY_4, UPGRADE_FSM_DUMMY_5, UPGRADE_PRE_CHECK_STARTED, UPGRADE_PRE_CHECK_IN_PROGRESS, UPGRADE_PRE_CHECK_SUCCESS.... Field introduced in 22.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	State *string `json:"state,omitempty"`
}
