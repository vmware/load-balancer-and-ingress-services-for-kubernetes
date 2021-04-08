package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudAutoscalingConfigFailureDetails cloud autoscaling config failure details
// swagger:model CloudAutoscalingConfigFailureDetails
type CloudAutoscalingConfigFailureDetails struct {

	// Cloud UUID. Field introduced in 20.1.1.
	CcID *string `json:"cc_id,omitempty"`

	// Failure reason if Autoscaling configuration fails. Field introduced in 20.1.1.
	ErrorString *string `json:"error_string,omitempty"`
}
