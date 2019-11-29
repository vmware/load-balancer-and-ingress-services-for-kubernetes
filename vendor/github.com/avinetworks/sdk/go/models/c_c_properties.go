package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CCProperties c c properties
// swagger:model CC_Properties
type CCProperties struct {

	// Number of rpc_poll_interval.
	RPCPollInterval *int32 `json:"rpc_poll_interval,omitempty"`

	// Number of rpc_queue_size.
	RPCQueueSize *int32 `json:"rpc_queue_size,omitempty"`
}
