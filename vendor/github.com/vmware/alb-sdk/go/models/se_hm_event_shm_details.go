package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

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
