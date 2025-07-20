// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// CCProperties c c properties
// swagger:model CC_Properties
type CCProperties struct {

	//  Unit is SEC. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RPCPollInterval *uint32 `json:"rpc_poll_interval,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RPCQueueSize *uint32 `json:"rpc_queue_size,omitempty"`
}
