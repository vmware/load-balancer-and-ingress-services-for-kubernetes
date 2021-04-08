package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// HTTPLocalFile HTTP local file
// swagger:model HTTPLocalFile
type HTTPLocalFile struct {

	// Mime-type of the content in the file.
	// Required: true
	ContentType *string `json:"content_type"`

	// File content to used in the local HTTP response body.
	// Required: true
	FileContent *string `json:"file_content"`
}
