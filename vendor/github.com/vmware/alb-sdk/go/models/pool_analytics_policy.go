// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PoolAnalyticsPolicy pool analytics policy
// swagger:model PoolAnalyticsPolicy
type PoolAnalyticsPolicy struct {

	// Enable real time metrics for server and pool metrics eg. l4_server.xxx, l7_server.xxx. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	EnableRealtimeMetrics *bool `json:"enable_realtime_metrics,omitempty"`
}
