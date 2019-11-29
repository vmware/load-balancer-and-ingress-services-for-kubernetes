package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ConfigActionDetails config action details
// swagger:model ConfigActionDetails
type ConfigActionDetails struct {

	// Name of the action.
	ActionName *string `json:"action_name,omitempty"`

	// Error message if request failed.
	ErrorMessage *string `json:"error_message,omitempty"`

	// Parameter data.
	ParameterData *string `json:"parameter_data,omitempty"`

	// API path.
	Path *string `json:"path,omitempty"`

	// Name of the resource.
	ResourceName *string `json:"resource_name,omitempty"`

	// Config type of the resource.
	ResourceType *string `json:"resource_type,omitempty"`

	// Status.
	Status *string `json:"status,omitempty"`

	// Request user.
	User *string `json:"user,omitempty"`
}
