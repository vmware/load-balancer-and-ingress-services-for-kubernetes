// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// FalsePositiveLearningConfig false positive learning config
// swagger:model FalsePositiveLearningConfig
type FalsePositiveLearningConfig struct {

	// Max number of applications supported to detect false positive. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MaxAppsSupported *int64 `json:"max_apps_supported,omitempty"`

	// Minimum monitor time required to automatically detect false positive. Unit is minutes. Field introduced in 22.1.1. Unit is MIN. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinMonitorTime *int64 `json:"min_monitor_time,omitempty"`

	// Minimum number of transactions in one application required to automatically detect false positive. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinTransPerApplication *int64 `json:"min_trans_per_application,omitempty"`

	// Minimum number of transactions in one URI required to automatically detect false positive. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MinTransPerURI *int64 `json:"min_trans_per_uri,omitempty"`
}
