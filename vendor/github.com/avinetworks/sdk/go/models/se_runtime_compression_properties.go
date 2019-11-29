package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeRuntimeCompressionProperties se runtime compression properties
// swagger:model SeRuntimeCompressionProperties
type SeRuntimeCompressionProperties struct {

	// If client RTT is higher than this threshold, enable normal compression on the response.
	MaxLowRtt *int32 `json:"max_low_rtt,omitempty"`

	// If client RTT is higher than this threshold, enable aggressive compression on the response.
	MinHighRtt *int32 `json:"min_high_rtt,omitempty"`

	// Minimum response content length to enable compression.
	MinLength *int32 `json:"min_length,omitempty"`

	// Values that identify mobile browsers in order to enable aggressive compression.
	MobileStr []string `json:"mobile_str,omitempty"`
}
