// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ImageEvent image event
// swagger:model ImageEvent
type ImageEvent struct {

	// Time taken to complete event in seconds. Field introduced in 21.1.3. Unit is SEC. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Duration *uint32 `json:"duration,omitempty"`

	// Task end time. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndTime *string `json:"end_time,omitempty"`

	// Ip of the node. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	IP *IPAddr `json:"ip,omitempty"`

	// Event message if any. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Message *string `json:"message,omitempty"`

	// Task start time. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StartTime *string `json:"start_time,omitempty"`

	// Event status. Enum options - SYSERR_SUCCESS, SYSERR_FAILURE, SYSERR_OUT_OF_MEMORY, SYSERR_NO_ENT, SYSERR_INVAL, SYSERR_ACCESS, SYSERR_FAULT, SYSERR_IO, SYSERR_TIMEOUT, SYSERR_NOT_SUPPORTED, SYSERR_NOT_READY, SYSERR_UPGRADE_IN_PROGRESS, SYSERR_WARM_START_IN_PROGRESS, SYSERR_TRY_AGAIN, SYSERR_NOT_UPGRADING, SYSERR_PENDING, SYSERR_EVENT_GEN_FAILURE, SYSERR_CONFIG_PARAM_MISSING, SYSERR_RANGE, SYSERR_BAD_REQUEST.... Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Status *string `json:"status,omitempty"`

	// Sub tasks executed on each node. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SubTasks []string `json:"sub_tasks,omitempty"`
}
