package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterServiceFailedEvent cluster service failed event
// swagger:model ClusterServiceFailedEvent
type ClusterServiceFailedEvent struct {

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`

	// Name of the controller service.
	ServiceName *string `json:"service_name,omitempty"`
}
