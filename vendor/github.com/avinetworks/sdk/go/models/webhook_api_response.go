package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// WebhookAPIResponse webhook Api response
// swagger:model WebhookApiResponse
type WebhookAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*Webhook `json:"results,omitempty"`
}
