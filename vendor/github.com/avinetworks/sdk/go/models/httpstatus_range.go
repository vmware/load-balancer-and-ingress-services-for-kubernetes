package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HttpstatusRange httpstatus range
// swagger:model HTTPStatusRange
type HttpstatusRange struct {

	// Starting HTTP response status code.
	// Required: true
	Begin *int32 `json:"begin"`

	// Ending HTTP response status code.
	// Required: true
	End *int32 `json:"end"`
}
