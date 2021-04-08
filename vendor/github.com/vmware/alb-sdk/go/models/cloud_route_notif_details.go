package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// CloudRouteNotifDetails cloud route notif details
// swagger:model CloudRouteNotifDetails
type CloudRouteNotifDetails struct {

	// Cloud id. Field introduced in 20.1.3.
	CcID *string `json:"cc_id,omitempty"`

	// Detailed reason for the route update notification. Field introduced in 20.1.3.
	Reason *string `json:"reason,omitempty"`

	// Name of route table for which update was performed. Field introduced in 20.1.3.
	RouteTable *string `json:"route_table,omitempty"`

	// Names of routes for which update was performed. Field introduced in 20.1.3.
	Routes []string `json:"routes,omitempty"`
}
