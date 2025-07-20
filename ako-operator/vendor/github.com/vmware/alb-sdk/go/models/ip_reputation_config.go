// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// IPReputationConfig Ip reputation config
// swagger:model IpReputationConfig
type IPReputationConfig struct {

	// IP reputation db file object expiry duration in days. Allowed values are 1-7. Field introduced in 20.1.1. Unit is DAYS. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IPReputationFileObjectExpiryDuration *uint32 `json:"ip_reputation_file_object_expiry_duration,omitempty"`

	// IP reputation db sync interval in minutes. Allowed values are 30-1440. Field introduced in 20.1.1. Unit is MIN. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 60), Basic edition(Allowed values- 60), Enterprise with Cloud Services edition.
	IPReputationSyncInterval *uint32 `json:"ip_reputation_sync_interval,omitempty"`
}
