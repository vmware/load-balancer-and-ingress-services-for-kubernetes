package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// VsMigrateParams vs migrate params
// swagger:model VsMigrateParams
type VsMigrateParams struct {

	//  It is a reference to an object of type ServiceEngine.
	FromSeRef *string `json:"from_se_ref,omitempty"`

	// Number of new_vcpus.
	NewVcpus *int32 `json:"new_vcpus,omitempty"`

	//  It is a reference to an object of type VIMgrHostRuntime.
	ToHostRef *string `json:"to_host_ref,omitempty"`

	// Placeholder for description of property to_new_se of obj type VsMigrateParams field type str  type boolean
	ToNewSe *bool `json:"to_new_se,omitempty"`

	//  It is a reference to an object of type ServiceEngine.
	ToSeRef *string `json:"to_se_ref,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`

	//  Field introduced in 17.1.1.
	// Required: true
	VipID *string `json:"vip_id"`
}
