package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AlertSyslogConfigAPIResponse alert syslog config Api response
// swagger:model AlertSyslogConfigApiResponse
type AlertSyslogConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*AlertSyslogConfig `json:"results,omitempty"`
}
