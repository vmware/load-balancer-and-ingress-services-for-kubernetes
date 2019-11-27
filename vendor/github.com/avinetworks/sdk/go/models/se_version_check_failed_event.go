package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeVersionCheckFailedEvent se version check failed event
// swagger:model SeVersionCheckFailedEvent
type SeVersionCheckFailedEvent struct {

	// Software version on the controller.
	ControllerVersion *string `json:"controller_version,omitempty"`

	// UUID of the SE.
	SeUUID *string `json:"se_uuid,omitempty"`

	// Software version on the SE.
	SeVersion *string `json:"se_version,omitempty"`
}
