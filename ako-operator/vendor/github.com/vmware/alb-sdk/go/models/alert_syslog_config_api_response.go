// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AlertSyslogConfigAPIResponse alert syslog config Api response
// swagger:model AlertSyslogConfigApiResponse
type AlertSyslogConfigAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*AlertSyslogConfig `json:"results,omitempty"`
}
