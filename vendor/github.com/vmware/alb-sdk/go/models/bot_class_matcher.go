// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotClassMatcher bot class matcher
// swagger:model BotClassMatcher
type BotClassMatcher struct {

	// The list of client classes. Enum options - UNDETERMINED_CLIENT, HUMAN_CLIENT, BOT_CLIENT. Field introduced in 21.1.1. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClientClasses []string `json:"client_classes,omitempty"`

	// The match operation. Enum options - IS_IN, IS_NOT_IN. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Op *string `json:"op,omitempty"`
}
