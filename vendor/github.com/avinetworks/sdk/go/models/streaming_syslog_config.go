package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// StreamingSyslogConfig streaming syslog config
// swagger:model StreamingSyslogConfig
type StreamingSyslogConfig struct {

	// Facility value, as defined in RFC5424, must be between 0 and 23 inclusive. Allowed values are 0-23. Field introduced in 18.1.1.
	Facility *int32 `json:"facility,omitempty"`

	// Severity code, as defined in RFC5424, for filtered logs. This must be between 0 and 7 inclusive. Allowed values are 0-7. Field introduced in 18.1.1.
	FilteredLogSeverity *int32 `json:"filtered_log_severity,omitempty"`

	// String to use as the hostname in the syslog messages. This *string can contain only printable ASCII characters (hex 21 to hex 7E; no space allowed). Field introduced in 18.1.1.
	Hostname *string `json:"hostname,omitempty"`

	// Severity code, as defined in RFC5424, for non-significant logs. This must be between 0 and 7 inclusive. Allowed values are 0-7. Field introduced in 18.1.1.
	NonSignificantLogSeverity *int32 `json:"non_significant_log_severity,omitempty"`

	// Severity code, as defined in RFC5424, for significant logs. This must be between 0 and 7 inclusive. Allowed values are 0-7. Field introduced in 18.1.1.
	SignificantLogSeverity *int32 `json:"significant_log_severity,omitempty"`
}
