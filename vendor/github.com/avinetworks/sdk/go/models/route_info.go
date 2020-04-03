package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// RouteInfo route info
// swagger:model RouteInfo
type RouteInfo struct {

	// Host interface name. Field introduced in 18.2.6.
	IfName *string `json:"if_name,omitempty"`

	// Network Namespace type and is used to add an route entry in a specific namespace. Enum options - LOCAL_NAMESPACE, HOST_NAMESPACE, OTHER_NAMESPACE. Field introduced in 18.2.6.
	NetworkNamespace *string `json:"network_namespace,omitempty"`

	// Host nexthop ip address. Field introduced in 18.2.6.
	Nexthop *IPAddr `json:"nexthop,omitempty"`

	// Host subnet address. Field introduced in 18.2.6.
	// Required: true
	Subnet *IPAddrPrefix `json:"subnet"`
}
