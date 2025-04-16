// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeHmEventShmDetails se hm event shm details
// swagger:model SeHmEventShmDetails
type SeHmEventShmDetails struct {

	// Average health monitor response time from server in milli-seconds. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AverageResponseTime *uint64 `json:"average_response_time,omitempty"`

	// Health Monitor name. It is a reference to an object of type HealthMonitor. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	HealthMonitor *string `json:"health_monitor"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RespString *string `json:"resp_string,omitempty"`

	// Response code from server. Field introduced in 17.2.4. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ResponseCode *uint32 `json:"response_code,omitempty"`
}
