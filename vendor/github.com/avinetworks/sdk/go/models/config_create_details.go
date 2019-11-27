package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConfigCreateDetails config create details
// swagger:model ConfigCreateDetails
type ConfigCreateDetails struct {

	// Error message if request failed.
	ErrorMessage *string `json:"error_message,omitempty"`

	// API path.
	Path *string `json:"path,omitempty"`

	// Request data if request failed.
	RequestData *string `json:"request_data,omitempty"`

	// Data of the created resource.
	ResourceData *string `json:"resource_data,omitempty"`

	// Name of the created resource.
	ResourceName *string `json:"resource_name,omitempty"`

	// Config type of the created resource.
	ResourceType *string `json:"resource_type,omitempty"`

	// Status.
	Status *string `json:"status,omitempty"`

	// Request user.
	User *string `json:"user,omitempty"`
}
