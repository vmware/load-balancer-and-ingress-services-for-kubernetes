package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RateLimiterProperties rate limiter properties
// swagger:model RateLimiterProperties
type RateLimiterProperties struct {

	// Number of stages in msf rate limiter. Allowed values are 1-2. Field introduced in 20.1.1.
	MsfNumStages *int32 `json:"msf_num_stages,omitempty"`

	// Each stage size in msf rate limiter. Field introduced in 20.1.1.
	MsfStageSize *int64 `json:"msf_stage_size,omitempty"`
}
