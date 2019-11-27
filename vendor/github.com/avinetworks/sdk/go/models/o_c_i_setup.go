package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OCISetup o c i setup
// swagger:model OCISetup
type OCISetup struct {

	// cc_id of OCISetup.
	CcID *string `json:"cc_id,omitempty"`

	// compartment_id of OCISetup.
	CompartmentID *string `json:"compartment_id,omitempty"`

	// reason of OCISetup.
	Reason *string `json:"reason,omitempty"`

	// status of OCISetup.
	Status *string `json:"status,omitempty"`

	// tenancy of OCISetup.
	Tenancy *string `json:"tenancy,omitempty"`

	// vcn_id of OCISetup.
	VcnID *string `json:"vcn_id,omitempty"`
}
