// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RecommendationRequest recommendation request
// swagger:model RecommendationRequest
type RecommendationRequest struct {

	// The match element for this a false positive should be mitigated. If this is not gives, all match elements will be considered. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	MatchElement *string `json:"match_element,omitempty"`

	// The report_timestamp field of the log entry. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	ReportTimestamp *string `json:"report_timestamp,omitempty"`

	// The request_id field of the log entry. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RequestID *string `json:"request_id,omitempty"`

	// The rule id for which a false positive should be mitigated. If this is not given, all rules will be considered. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	RuleID *string `json:"rule_id,omitempty"`

	// The type of the request, e.g. RECOMMENDATION_REQUEST_FALSE_POSITIVE. Enum options - RECOMMENDATION_REQUEST_FALSE_POSITIVE. Field introduced in 21.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	// Required: true
	Type *string `json:"type"`
}
