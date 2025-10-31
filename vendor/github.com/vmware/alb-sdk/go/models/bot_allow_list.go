// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotAllowList bot allow list
// swagger:model BotAllowList
type BotAllowList struct {

	// Allow rules to control which requests undergo BOT detection. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Rules []*BotAllowRule `json:"rules,omitempty"`
}
