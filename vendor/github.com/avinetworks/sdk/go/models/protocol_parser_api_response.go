package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ProtocolParserAPIResponse protocol parser Api response
// swagger:model ProtocolParserApiResponse
type ProtocolParserAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*ProtocolParser `json:"results,omitempty"`
}
