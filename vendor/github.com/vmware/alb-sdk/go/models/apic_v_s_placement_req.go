package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// ApicVSPlacementReq apic v s placement req
// swagger:model ApicVSPlacementReq
type ApicVSPlacementReq struct {

	// graph of ApicVSPlacementReq.
	Graph *string `json:"graph,omitempty"`

	// Placeholder for description of property lifs of obj type ApicVSPlacementReq field type str  type object
	Lifs []*Lif `json:"lifs,omitempty"`

	// Placeholder for description of property network_rel of obj type ApicVSPlacementReq field type str  type object
	NetworkRel []*APICNetworkRel `json:"network_rel,omitempty"`

	// tenant_name of ApicVSPlacementReq.
	TenantName *string `json:"tenant_name,omitempty"`

	// vdev of ApicVSPlacementReq.
	Vdev *string `json:"vdev,omitempty"`

	// vgrp of ApicVSPlacementReq.
	Vgrp *string `json:"vgrp,omitempty"`
}
