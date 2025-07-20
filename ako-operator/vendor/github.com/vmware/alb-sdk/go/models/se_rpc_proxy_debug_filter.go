// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeRPCProxyDebugFilter se Rpc proxy debug filter
// swagger:model SeRpcProxyDebugFilter
type SeRPCProxyDebugFilter struct {

	// Method name of RPC. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	MethodName *string `json:"method_name,omitempty"`

	// Queue name of RPC. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	Queue *string `json:"queue,omitempty"`

	// UUID of Service Engine. Field introduced in 18.1.5, 18.2.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	SeUUID *string `json:"se_uuid,omitempty"`
}
