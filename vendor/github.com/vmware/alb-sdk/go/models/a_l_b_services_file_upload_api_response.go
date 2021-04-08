package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ALBServicesFileUploadAPIResponse a l b services file upload Api response
// swagger:model ALBServicesFileUploadApiResponse
type ALBServicesFileUploadAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// next
	Next *string `json:"next,omitempty"`

	// results
	// Required: true
	Results []*ALBServicesFileUpload `json:"results,omitempty"`
}
