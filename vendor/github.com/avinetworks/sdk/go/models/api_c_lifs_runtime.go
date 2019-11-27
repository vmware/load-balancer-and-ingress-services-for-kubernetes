package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// APICLifsRuntime API c lifs runtime
// swagger:model APICLifsRuntime
type APICLifsRuntime struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Placeholder for description of property auto_allocated of obj type APICLifsRuntime field type str  type boolean
	AutoAllocated *bool `json:"auto_allocated,omitempty"`

	// Placeholder for description of property cifs of obj type APICLifsRuntime field type str  type object
	Cifs []*Cif `json:"cifs,omitempty"`

	// Contract Graph associated with the VirtualService. Field introduced in 17.2.14,18.1.5,18.2.1.
	ContractGraphs []string `json:"contract_graphs,omitempty"`

	// lif_label of APICLifsRuntime.
	// Required: true
	LifLabel *string `json:"lif_label"`

	// Placeholder for description of property multi_vrf of obj type APICLifsRuntime field type str  type boolean
	MultiVrf *bool `json:"multi_vrf,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// network of APICLifsRuntime.
	Network *string `json:"network,omitempty"`

	// Unique object identifier of se.
	SeUUID []string `json:"se_uuid,omitempty"`

	// subnet of APICLifsRuntime.
	Subnet *string `json:"subnet,omitempty"`

	// tenant_name of APICLifsRuntime.
	// Required: true
	TenantName *string `json:"tenant_name"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`

	// Unique object identifier of transaction.
	TransactionUUID []string `json:"transaction_uuid,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Unique object identifier of vs.
	VsUUID []string `json:"vs_uuid,omitempty"`
}
