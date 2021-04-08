package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotDetectionPolicyAPIResponse bot detection policy Api response
// swagger:model BotDetectionPolicyApiResponse
type BotDetectionPolicyAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*BotDetectionPolicy `json:"results,omitempty"`
}
