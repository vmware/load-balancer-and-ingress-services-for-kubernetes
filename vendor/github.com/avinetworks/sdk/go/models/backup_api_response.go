package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BackupAPIResponse backup Api response
// swagger:model BackupApiResponse
type BackupAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*Backup `json:"results,omitempty"`
}
