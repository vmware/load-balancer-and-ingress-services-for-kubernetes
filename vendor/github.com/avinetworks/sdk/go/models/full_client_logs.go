package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FullClientLogs full client logs
// swagger:model FullClientLogs
type FullClientLogs struct {

	// [DEPRECATED] Log all headers. Please use the all_headers flag in AnalyticsPolicy. Field deprecated in 18.1.4, 18.2.1.
	AllHeaders *bool `json:"all_headers,omitempty"`

	// How long should the system capture all logs, measured in minutes. Set to 0 for infinite. Special values are 0 - 'infinite'.
	Duration *int32 `json:"duration,omitempty"`

	// Capture all client logs including connections and requests.  When disabled, only errors will be logged.
	// Required: true
	Enabled *bool `json:"enabled"`

	// This setting limits the number of non-significant logs generated per second for this VS on each SE. Default is 10 logs per second. Set it to zero (0) to disable throttling. Field introduced in 17.1.3.
	Throttle *int32 `json:"throttle,omitempty"`
}
