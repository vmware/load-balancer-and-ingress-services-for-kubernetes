package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConfigUpdateDetails config update details
// swagger:model ConfigUpdateDetails
type ConfigUpdateDetails struct {

	// Error message if request failed.
	ErrorMessage *string `json:"error_message,omitempty"`

	// New updated data of the resource.
	NewResourceData *string `json:"new_resource_data,omitempty"`

	// Old & overwritten data of the resource.
	OldResourceData *string `json:"old_resource_data,omitempty"`

	// API path.
	Path *string `json:"path,omitempty"`

	// Request data if request failed.
	RequestData *string `json:"request_data,omitempty"`

	// Name of the created resource.
	ResourceName *string `json:"resource_name,omitempty"`

	// Config type of the updated resource.
	ResourceType *string `json:"resource_type,omitempty"`

	// Status.
	Status *string `json:"status,omitempty"`

	// Request user.
	User *string `json:"user,omitempty"`
}
