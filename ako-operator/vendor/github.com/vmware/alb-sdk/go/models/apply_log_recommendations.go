// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// ApplyLogRecommendations apply log recommendations
// swagger:model ApplyLogRecommendations
type ApplyLogRecommendations struct {

	// Describe the actions we want to perform. Field introduced in 21.1.3. Minimum of 1 items required. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	Actions []*Action `json:"actions,omitempty"`
}
