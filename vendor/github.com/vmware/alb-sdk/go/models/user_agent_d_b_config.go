// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// UserAgentDBConfig user agent d b config
// swagger:model UserAgentDBConfig
type UserAgentDBConfig struct {

	// Batch query limit. Allowed values are 1-500. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Essentials edition(Allowed values- 500), Basic edition(Allowed values- 500), Enterprise with Cloud Services edition.
	AllowedBatchSize *uint32 `json:"allowed_batch_size,omitempty"`
}
