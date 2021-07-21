// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// SeRPCProxyDebugFilter se Rpc proxy debug filter
// swagger:model SeRpcProxyDebugFilter
type SeRPCProxyDebugFilter struct {

	// Method name of RPC. Field introduced in 18.1.5, 18.2.1.
	MethodName *string `json:"method_name,omitempty"`

	// Queue name of RPC. Field introduced in 18.1.5, 18.2.1.
	Queue *string `json:"queue,omitempty"`

	// UUID of Service Engine. Field introduced in 18.1.5, 18.2.1.
	SeUUID *string `json:"se_uuid,omitempty"`
}
