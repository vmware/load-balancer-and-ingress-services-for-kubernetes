package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// GCPSetup g c p setup
// swagger:model GCPSetup
type GCPSetup struct {

	// cc_id of GCPSetup.
	CcID *string `json:"cc_id,omitempty"`

	// hostname of GCPSetup.
	Hostname *string `json:"hostname,omitempty"`

	// network of GCPSetup.
	Network *string `json:"network,omitempty"`

	// nhop_inst of GCPSetup.
	NhopInst *string `json:"nhop_inst,omitempty"`

	// Placeholder for description of property nhop_ip of obj type GCPSetup field type str  type object
	NhopIP *IPAddr `json:"nhop_ip,omitempty"`

	// project of GCPSetup.
	Project *string `json:"project,omitempty"`

	// reason of GCPSetup.
	Reason *string `json:"reason,omitempty"`

	// route_name of GCPSetup.
	RouteName *string `json:"route_name,omitempty"`

	// subnet of GCPSetup.
	Subnet *string `json:"subnet,omitempty"`

	// Placeholder for description of property vip of obj type GCPSetup field type str  type object
	Vip *IPAddr `json:"vip,omitempty"`

	// vs_name of GCPSetup.
	VsName *string `json:"vs_name,omitempty"`

	// Unique object identifier of vs.
	VsUUID *string `json:"vs_uuid,omitempty"`

	// zone of GCPSetup.
	Zone *string `json:"zone,omitempty"`
}
