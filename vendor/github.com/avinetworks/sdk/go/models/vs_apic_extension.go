package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsApicExtension vs apic extension
// swagger:model VsApicExtension
type VsApicExtension struct {

	// Unique object identifier of se.
	SeUUID *string `json:"se_uuid,omitempty"`

	// Unique object identifier of txn.
	// Required: true
	TxnUUID *string `json:"txn_uuid"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	// Placeholder for description of property vnic of obj type VsApicExtension field type str  type object
	Vnic []*VsSeVnic `json:"vnic,omitempty"`
}
