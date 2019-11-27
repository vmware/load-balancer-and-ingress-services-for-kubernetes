package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HttpstatusMatch httpstatus match
// swagger:model HTTPStatusMatch
type HttpstatusMatch struct {

	// Criterion to use for matching the HTTP response status code(s). Enum options - IS_IN, IS_NOT_IN.
	// Required: true
	MatchCriteria *string `json:"match_criteria"`

	// HTTP response status code range(s).
	Ranges []*HttpstatusRange `json:"ranges,omitempty"`

	// HTTP response status code(s).
	StatusCodes []int64 `json:"status_codes,omitempty,omitempty"`
}
