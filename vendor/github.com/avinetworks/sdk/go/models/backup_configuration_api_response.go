package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BackupConfigurationAPIResponse backup configuration Api response
// swagger:model BackupConfigurationApiResponse
type BackupConfigurationAPIResponse struct {

	// count
	// Required: true
	Count *int32 `json:"count"`

	// results
	// Required: true
	Results []*BackupConfiguration `json:"results,omitempty"`
}
