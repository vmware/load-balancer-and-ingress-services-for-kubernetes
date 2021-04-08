package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RateLimiterEventDetails rate limiter event details
// swagger:model RateLimiterEventDetails
type RateLimiterEventDetails struct {

	// Rate limiter error message. Field introduced in 20.1.1.
	ErrorMessage *string `json:"error_message,omitempty"`

	// Name of the rate limiter. Field introduced in 20.1.1.
	RlResourceName *string `json:"rl_resource_name,omitempty"`

	// Rate limiter type. Field introduced in 20.1.1.
	RlResourceType *string `json:"rl_resource_type,omitempty"`

	// Status. Field introduced in 20.1.1.
	Status *string `json:"status,omitempty"`
}
