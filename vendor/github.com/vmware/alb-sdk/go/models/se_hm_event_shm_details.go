// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHmEventShmDetails se hm event shm details
// swagger:model SeHmEventShmDetails
type SeHmEventShmDetails struct {

	// Average health monitor response time from server in milli-seconds.
	AverageResponseTime *int64 `json:"average_response_time,omitempty"`

	// Health Monitor name. It is a reference to an object of type HealthMonitor.
	// Required: true
	HealthMonitor *string `json:"health_monitor"`

	// resp_string of SeHmEventShmDetails.
	RespString *string `json:"resp_string,omitempty"`

	// Response code from server. Field introduced in 17.2.4.
	ResponseCode *int32 `json:"response_code,omitempty"`
}
