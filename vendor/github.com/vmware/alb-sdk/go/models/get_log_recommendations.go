// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// GetLogRecommendations get log recommendations
// swagger:model GetLogRecommendations
type GetLogRecommendations struct {

	// Describe the recommendation we want to get. Field introduced in 21.1.3. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Requests []*RecommendationRequest `json:"requests,omitempty"`
}
