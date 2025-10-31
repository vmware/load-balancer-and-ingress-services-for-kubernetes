// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PulseServicesSessionConfig pulse services session config
// swagger:model PulseServicesSessionConfig
type PulseServicesSessionConfig struct {

	// Session Headers. Field introduced in 30.2.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	SessionHeaders []*SessionHeaders `json:"session_headers,omitempty"`
}
