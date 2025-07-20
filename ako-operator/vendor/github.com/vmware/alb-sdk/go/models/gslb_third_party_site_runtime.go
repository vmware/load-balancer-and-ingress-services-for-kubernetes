// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GslbThirdPartySiteRuntime gslb third party site runtime
// swagger:model GslbThirdPartySiteRuntime
type GslbThirdPartySiteRuntime struct {

	// This field will provide information on origin(site name) of the health monitoring information. Field introduced in 22.1.5. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	HealthMonitorInfo *string `json:"health_monitor_info,omitempty"`

	//  Field introduced in 17.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SiteInfo *GslbSiteRuntimeInfo `json:"site_info,omitempty"`
}
