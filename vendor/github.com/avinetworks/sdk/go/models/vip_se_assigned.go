package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VipSeAssigned vip se assigned
// swagger:model VipSeAssigned
type VipSeAssigned struct {

	// Placeholder for description of property admin_down_requested of obj type VipSeAssigned field type str  type boolean
	AdminDownRequested *bool `json:"admin_down_requested,omitempty"`

	// Placeholder for description of property connected of obj type VipSeAssigned field type str  type boolean
	Connected *bool `json:"connected,omitempty"`

	// Name of the object.
	Name *string `json:"name,omitempty"`

	// Placeholder for description of property oper_status of obj type VipSeAssigned field type str  type object
	OperStatus *OperationalStatus `json:"oper_status,omitempty"`

	// Placeholder for description of property primary of obj type VipSeAssigned field type str  type boolean
	Primary *bool `json:"primary,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	Ref *string `json:"ref,omitempty"`

	// Placeholder for description of property scalein_in_progress of obj type VipSeAssigned field type str  type boolean
	ScaleinInProgress *bool `json:"scalein_in_progress,omitempty"`

	// Placeholder for description of property snat_ip of obj type VipSeAssigned field type str  type object
	SnatIP *IPAddr `json:"snat_ip,omitempty"`

	// Placeholder for description of property standby of obj type VipSeAssigned field type str  type boolean
	Standby *bool `json:"standby,omitempty"`
}
