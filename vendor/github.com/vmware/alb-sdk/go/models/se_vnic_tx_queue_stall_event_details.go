package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// SeVnicTxQueueStallEventDetails se vnic tx queue stall event details
// swagger:model SeVnicTxQueueStallEventDetails
type SeVnicTxQueueStallEventDetails struct {

	// Vnic name.
	IfName *string `json:"if_name,omitempty"`

	// Vnic Linux name.
	LinuxName *string `json:"linux_name,omitempty"`

	// Queue number.
	Queue *int32 `json:"queue,omitempty"`

	// UUID of the SE responsible for this event. It is a reference to an object of type ServiceEngine.
	SeRef *string `json:"se_ref,omitempty"`
}
