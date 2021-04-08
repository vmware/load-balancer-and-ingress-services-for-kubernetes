package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsScaleoutParams vs scaleout params
// swagger:model VsScaleoutParams
type VsScaleoutParams struct {

	// Placeholder for description of property admin_up of obj type VsScaleoutParams field type str  type boolean
	AdminUp *bool `json:"admin_up,omitempty"`

	// Number of new_vcpus.
	NewVcpus *int32 `json:"new_vcpus,omitempty"`

	//  It is a reference to an object of type VIMgrHostRuntime.
	ToHostRef *string `json:"to_host_ref,omitempty"`

	// Placeholder for description of property to_new_se of obj type VsScaleoutParams field type str  type boolean
	ToNewSe *bool `json:"to_new_se,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	ToSeRef *string `json:"to_se_ref,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	//  Field introduced in 17.1.1.
	// Required: true
	VipID *string `json:"vip_id"`
}
