// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// PerformanceLimits performance limits
// swagger:model PerformanceLimits
type PerformanceLimits struct {

	// The maximum number of concurrent client conections allowed to the Virtual Service.
	MaxConcurrentConnections *int32 `json:"max_concurrent_connections,omitempty"`

	// The maximum throughput per second for all clients allowed through the client side of the Virtual Service.
	MaxThroughput *int32 `json:"max_throughput,omitempty"`
}
