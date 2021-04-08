package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// FailActionHTTPLocalResponse fail action HTTP local response
// swagger:model FailActionHTTPLocalResponse
type FailActionHTTPLocalResponse struct {

	// Placeholder for description of property file of obj type FailActionHTTPLocalResponse field type str  type object
	File *HTTPLocalFile `json:"file,omitempty"`

	//  Enum options - FAIL_HTTP_STATUS_CODE_200, FAIL_HTTP_STATUS_CODE_503.
	StatusCode *string `json:"status_code,omitempty"`
}
