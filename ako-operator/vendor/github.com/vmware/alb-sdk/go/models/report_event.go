// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ReportEvent report event
// swagger:model ReportEvent
type ReportEvent struct {

	// Time taken to complete event in seconds. Field introduced in 22.1.6, 30.2.1. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Duration *uint32 `json:"duration,omitempty"`

	// Event end time. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndTime *string `json:"end_time,omitempty"`

	// Name representing the event. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EventName *string `json:"event_name,omitempty"`

	// Event message if any. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Messages []string `json:"messages,omitempty"`

	// Event start time. Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StartTime *string `json:"start_time,omitempty"`

	// Event status. Enum options - SYSERR_SUCCESS, SYSERR_FAILURE, SYSERR_OUT_OF_MEMORY, SYSERR_NO_ENT, SYSERR_INVAL, SYSERR_ACCESS, SYSERR_FAULT, SYSERR_IO, SYSERR_TIMEOUT, SYSERR_NOT_SUPPORTED, SYSERR_NOT_READY, SYSERR_UPGRADE_IN_PROGRESS, SYSERR_WARM_START_IN_PROGRESS, SYSERR_TRY_AGAIN, SYSERR_NOT_UPGRADING, SYSERR_PENDING, SYSERR_EVENT_GEN_FAILURE, SYSERR_CONFIG_PARAM_MISSING, SYSERR_RANGE, SYSERR_BAD_REQUEST.... Field introduced in 22.1.6, 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Status *string `json:"status,omitempty"`
}
