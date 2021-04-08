package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VcenterHosts vcenter hosts
// swagger:model VcenterHosts
type VcenterHosts struct {

	//  It is a reference to an object of type VIMgrHostRuntime.
	HostRefs []string `json:"host_refs,omitempty"`

	// Placeholder for description of property include of obj type VcenterHosts field type str  type boolean
	Include *bool `json:"include,omitempty"`
}
