// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// BotTypeMatcher bot type matcher
// swagger:model BotTypeMatcher
type BotTypeMatcher struct {

	// The list of client types. Enum options - UNDETERMINED_CLIENT_TYPE, WEB_BROWSER, IN_APP_BROWSER, SEARCH_ENGINE, IMPERSONATOR, SPAM_SOURCE, WEB_ATTACKS, BOTNET, SCANNER, DENIAL_OF_SERVICE, CLOUD_SOURCE, SECURITY_SCANNER, SITE_MONITOR, GENERIC_APPLICATION, SUSPICIOUS_APPLICATION. Field introduced in 21.1.1. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ClientTypes []string `json:"client_types,omitempty"`

	// The match operation. Enum options - IS_IN, IS_NOT_IN. Field introduced in 21.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Op *string `json:"op,omitempty"`
}
