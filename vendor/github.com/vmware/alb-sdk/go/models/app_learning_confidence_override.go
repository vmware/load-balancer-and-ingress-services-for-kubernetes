// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AppLearningConfidenceOverride app learning confidence override
// swagger:model AppLearningConfidenceOverride
type AppLearningConfidenceOverride struct {

	// Confidence threshold for label CONFIDENCE_HIGH. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfidHighValue *uint32 `json:"confid_high_value,omitempty"`

	// Confidence threshold for label CONFIDENCE_LOW. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfidLowValue *uint32 `json:"confid_low_value,omitempty"`

	// Confidence threshold for label CONFIDENCE_PROBABLE. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfidProbableValue *uint32 `json:"confid_probable_value,omitempty"`

	// Confidence threshold for label CONFIDENCE_VERY_HIGH. Field introduced in 18.2.3. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	ConfidVeryHighValue *uint32 `json:"confid_very_high_value,omitempty"`
}
