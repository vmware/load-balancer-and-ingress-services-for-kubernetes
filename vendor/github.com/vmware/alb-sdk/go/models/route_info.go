// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// RouteInfo route info
// swagger:model RouteInfo
type RouteInfo struct {

	// Host interface name. Field introduced in 18.2.6.
	IfName *string `json:"if_name,omitempty"`

	// Network Namespace type used to add an route entry in a specific namespace. Enum options - LOCAL_NAMESPACE, HOST_NAMESPACE, OTHER_NAMESPACE. Field introduced in 18.2.6.
	NetworkNamespace *string `json:"network_namespace,omitempty"`

	// Host nexthop ip address. Field introduced in 18.2.6.
	Nexthop *IPAddr `json:"nexthop,omitempty"`

	// Host subnet address. Field introduced in 18.2.6.
	// Required: true
	Subnet *IPAddrPrefix `json:"subnet"`
}
