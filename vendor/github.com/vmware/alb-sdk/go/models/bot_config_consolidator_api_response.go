package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BotConfigConsolidatorAPIResponse bot config consolidator Api response
// swagger:model BotConfigConsolidatorApiResponse
type BotConfigConsolidatorAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*BotConfigConsolidator `json:"results,omitempty"`
}
