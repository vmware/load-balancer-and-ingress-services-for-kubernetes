package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ClusterServiceCriticalFailureEvent cluster service critical failure event
// swagger:model ClusterServiceCriticalFailureEvent
type ClusterServiceCriticalFailureEvent struct {

	// Name of controller node.
	NodeName *string `json:"node_name,omitempty"`

	// Name of the controller service.
	ServiceName *string `json:"service_name,omitempty"`
}
