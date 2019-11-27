package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// StaticRoute static route
// swagger:model StaticRoute
type StaticRoute struct {

	// Disable the gateway monitor for default gateway. They are monitored by default. Field introduced in 17.1.1.
	DisableGatewayMonitor *bool `json:"disable_gateway_monitor,omitempty"`

	// if_name of StaticRoute.
	IfName *string `json:"if_name,omitempty"`

	// Placeholder for description of property next_hop of obj type StaticRoute field type str  type object
	// Required: true
	NextHop *IPAddr `json:"next_hop"`

	// Placeholder for description of property prefix of obj type StaticRoute field type str  type object
	// Required: true
	Prefix *IPAddrPrefix `json:"prefix"`

	// route_id of StaticRoute.
	// Required: true
	RouteID *string `json:"route_id"`
}
