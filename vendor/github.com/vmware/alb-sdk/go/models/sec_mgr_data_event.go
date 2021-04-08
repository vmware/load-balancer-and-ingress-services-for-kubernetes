package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SecMgrDataEvent sec mgr data event
// swagger:model SecMgrDataEvent
type SecMgrDataEvent struct {

	// Type of the generated for an application. Field introduced in 20.1.1.
	Error *string `json:"error,omitempty"`

	// Name of the application. Field introduced in 20.1.1.
	Name *string `json:"name,omitempty"`
}
