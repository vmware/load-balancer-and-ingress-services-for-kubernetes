// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ProtocolMatch protocol match
// swagger:model ProtocolMatch
type ProtocolMatch struct {

	// Criterion to use for protocol matching the HTTP request. Enum options - IS_IN, IS_NOT_IN.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// HTTP or HTTPS protocol. Enum options - HTTP, HTTPS.
	// Required: true
	Protocols *string `json:"protocols"`
}
