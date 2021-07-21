// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CCProperties c c properties
// swagger:model CC_Properties
type CCProperties struct {

	//  Unit is SEC.
	RPCPollInterval *int32 `json:"rpc_poll_interval,omitempty"`

	// Number of rpc_queue_size.
	RPCQueueSize *int32 `json:"rpc_queue_size,omitempty"`
}
