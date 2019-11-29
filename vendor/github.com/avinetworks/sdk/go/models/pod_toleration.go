package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PodToleration pod toleration
// swagger:model PodToleration
type PodToleration struct {

	// Effect to match. Enum options - NO_SCHEDULE, PREFER_NO_SCHEDULE, NO_EXECUTE. Field introduced in 17.2.14, 18.1.5, 18.2.1.
	Effect *string `json:"effect,omitempty"`

	// Key to match. Field introduced in 17.2.14, 18.1.5, 18.2.1.
	Key *string `json:"key,omitempty"`

	// Operator to match. Enum options - EQUAL, EXISTS. Field introduced in 17.2.14, 18.1.5, 18.2.1.
	Operator *string `json:"operator,omitempty"`

	// Pods that tolerate the taint with a specified toleration_seconds remain bound for the specified amount of time. Field introduced in 17.2.14, 18.1.5, 18.2.1.
	TolerationSeconds *int32 `json:"toleration_seconds,omitempty"`

	// Value to match. Field introduced in 17.2.14, 18.1.5, 18.2.1.
	Value *string `json:"value,omitempty"`
}
