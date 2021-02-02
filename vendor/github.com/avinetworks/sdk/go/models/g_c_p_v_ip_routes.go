package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GCPVIPRoutes g c p v IP routes
// swagger:model GCPVIPRoutes
type GCPVIPRoutes struct {

	// Match SE group subnets for VIP placement. Default is to not match SE group subnets. Field introduced in 18.2.9, 20.1.1.
	MatchSeGroupSubnet *bool `json:"match_se_group_subnet,omitempty"`
}
