package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OpenStackRoleMapping open stack role mapping
// swagger:model OpenStackRoleMapping
type OpenStackRoleMapping struct {

	// Role name in Avi.
	// Required: true
	AviRole *string `json:"avi_role"`

	// Role name in OpenStack.
	// Required: true
	OsRole *string `json:"os_role"`
}
