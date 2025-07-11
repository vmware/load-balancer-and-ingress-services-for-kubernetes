// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// StreamingSyslogConfig streaming syslog config
// swagger:model StreamingSyslogConfig
type StreamingSyslogConfig struct {

	// Facility value, as defined in RFC5424, must be between 0 and 23 inclusive. Allowed values are 0-23. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Facility *uint32 `json:"facility,omitempty"`

	// Severity code, as defined in RFC5424, for filtered logs. This must be between 0 and 7 inclusive. Allowed values are 0-7. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	FilteredLogSeverity *uint32 `json:"filtered_log_severity,omitempty"`

	// String to use as the hostname in the syslog messages. This *string can contain only printable ASCII characters (hex 21 to hex 7E; no space allowed). Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Hostname *string `json:"hostname,omitempty"`

	// As per RFC, Constant *string to identify the type of message. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MsgID *string `json:"msg_id,omitempty"`

	// Severity code, as defined in RFC5424, for non-significant logs. This must be between 0 and 7 inclusive. Allowed values are 0-7. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	NonSignificantLogSeverity *uint32 `json:"non_significant_log_severity,omitempty"`

	// As per RFC, if there is a change in value indicated there has been discontinuity in syslog reporting. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ProcID *string `json:"proc_id,omitempty"`

	// Severity code, as defined in RFC5424, for significant logs. This must be between 0 and 7 inclusive. Allowed values are 0-7. Field introduced in 18.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SignificantLogSeverity *uint32 `json:"significant_log_severity,omitempty"`
}
