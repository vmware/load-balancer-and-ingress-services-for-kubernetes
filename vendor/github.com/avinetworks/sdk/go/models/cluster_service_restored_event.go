package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterServiceRestoredEvent cluster service restored event
// swagger:model ClusterServiceRestoredEvent
type ClusterServiceRestoredEvent struct {

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`

	// Name of the controller service.
	ServiceName *string `json:"service_name,omitempty"`
}
