// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// MarathonServicePortConflict marathon service port conflict
// swagger:model MarathonServicePortConflict
type MarathonServicePortConflict struct {

	// app_name of MarathonServicePortConflict.
	AppName *string `json:"app_name,omitempty"`

	// cc_id of MarathonServicePortConflict.
	CcID *string `json:"cc_id,omitempty"`

	// marathon_url of MarathonServicePortConflict.
	// Required: true
	MarathonURL *string `json:"marathon_url"`

	// Number of port.
	// Required: true
	Port *int32 `json:"port"`
}
