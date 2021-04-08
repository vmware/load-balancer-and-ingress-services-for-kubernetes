package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// BgpRoutingOptions bgp routing options
// swagger:model BgpRoutingOptions
type BgpRoutingOptions struct {

	// Advertise self as default router to the BGP peer. Field introduced in 20.1.1.
	AdvertiseDefaultRoute *bool `json:"advertise_default_route,omitempty"`

	// Advertise the learned routes to the BGP peer. Field introduced in 20.1.1.
	AdvertiseLearnedRoutes *bool `json:"advertise_learned_routes,omitempty"`

	// Features are applied to peers matching this label. Field introduced in 20.1.1.
	// Required: true
	Label *string `json:"label"`

	// Learn only default route from the BGP peer. Field introduced in 20.1.1.
	LearnOnlyDefaultRoute *bool `json:"learn_only_default_route,omitempty"`

	// Learn routes from the BGP peer. Field introduced in 20.1.1.
	LearnRoutes *bool `json:"learn_routes,omitempty"`

	// Maximum number of routes that can be learned from a BGP peer. Allowed values are 50-250. Field introduced in 20.1.1.
	MaxLearnLimit *int32 `json:"max_learn_limit,omitempty"`
}
