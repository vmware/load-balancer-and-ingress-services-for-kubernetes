package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// OpenstackLogin openstack login
// swagger:model OpenstackLogin
type OpenstackLogin struct {

	// admin_tenant of OpenstackLogin.
	AdminTenant *string `json:"admin_tenant,omitempty"`

	// auth_url of OpenstackLogin.
	AuthURL *string `json:"auth_url,omitempty"`

	// keystone_host of OpenstackLogin.
	KeystoneHost *string `json:"keystone_host,omitempty"`

	// password of OpenstackLogin.
	// Required: true
	Password *string `json:"password"`

	// region of OpenstackLogin.
	Region *string `json:"region,omitempty"`

	// username of OpenstackLogin.
	// Required: true
	Username *string `json:"username"`
}
