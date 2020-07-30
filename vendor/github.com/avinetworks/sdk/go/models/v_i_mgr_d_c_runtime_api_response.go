package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VIMgrDCRuntimeAPIResponse v i mgr d c runtime Api response
// swagger:model VIMgrDCRuntimeApiResponse
type VIMgrDCRuntimeAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*VIMgrDCRuntime `json:"results,omitempty"`
}
