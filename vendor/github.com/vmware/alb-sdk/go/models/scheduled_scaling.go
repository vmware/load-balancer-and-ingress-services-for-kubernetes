// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ScheduledScaling scheduled scaling
// swagger:model ScheduledScaling
type ScheduledScaling struct {

	// Scheduled autoscale duration (in hours). Allowed values are 1-24. Field introduced in 21.1.1. Unit is HOURS. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AutoscalingDuration *uint32 `json:"autoscaling_duration,omitempty"`

	// The cron expression describing desired time for the scheduled autoscale. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CronExpression *string `json:"cron_expression,omitempty"`

	// Desired number of servers during scheduled intervals, it may cause scale-in or scale-out based on the value. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DesiredCapacity *uint32 `json:"desired_capacity,omitempty"`

	// Enables the scheduled autoscale. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Enable *bool `json:"enable,omitempty"`

	// Scheduled autoscale end date in ISO8601 format, said day will be included in scheduled and have to be in future and greater than start date. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	EndDate *string `json:"end_date,omitempty"`

	// Maximum number of simultaneous scale-in/out servers for scheduled autoscale. If this value is 0, regular autoscale policy dictates this. . Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ScheduleMaxStep *uint32 `json:"schedule_max_step,omitempty"`

	// Scheduled autoscale start date in ISO8601 format, said day will be included in scheduled and have to be in future. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	StartDate *string `json:"start_date,omitempty"`
}
