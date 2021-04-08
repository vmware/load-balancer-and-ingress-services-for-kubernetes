package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AppLearningConfidenceOverride app learning confidence override
// swagger:model AppLearningConfidenceOverride
type AppLearningConfidenceOverride struct {

	// Confidence threshold for label CONFIDENCE_HIGH. Field introduced in 18.2.3.
	ConfidHighValue *int32 `json:"confid_high_value,omitempty"`

	// Confidence threshold for label CONFIDENCE_LOW. Field introduced in 18.2.3.
	ConfidLowValue *int32 `json:"confid_low_value,omitempty"`

	// Confidence threshold for label CONFIDENCE_PROBABLE. Field introduced in 18.2.3.
	ConfidProbableValue *int32 `json:"confid_probable_value,omitempty"`

	// Confidence threshold for label CONFIDENCE_VERY_HIGH. Field introduced in 18.2.3.
	ConfidVeryHighValue *int32 `json:"confid_very_high_value,omitempty"`
}
