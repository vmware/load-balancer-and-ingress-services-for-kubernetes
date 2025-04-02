// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ControllerInfo controller info
// swagger:model ControllerInfo
type ControllerInfo struct {

	// Total controller memory usage in GBs. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	CurrentControllerMemUsage *float64 `json:"current_controller_mem_usage,omitempty"`
}
