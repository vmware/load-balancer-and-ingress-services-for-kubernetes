package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// PortalFileUploadAPIResponse portal file upload Api response
// swagger:model PortalFileUploadApiResponse
type PortalFileUploadAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*PortalFileUpload `json:"results,omitempty"`
}
