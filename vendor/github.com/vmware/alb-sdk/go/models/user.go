package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// User user
// swagger:model User
type User struct {

	// UNIX time since epoch in microseconds. Units(MICROSECONDS).
	// Read Only: true
	LastModified *string `json:"_last_modified,omitempty"`

	// Placeholder for description of property access of obj type User field type str  type object
	Access []*UserRole `json:"access,omitempty"`

	//  It is a reference to an object of type Tenant.
	DefaultTenantRef *string `json:"default_tenant_ref,omitempty"`

	// email of User.
	Email *string `json:"email,omitempty"`

	// full_name of User.
	FullName *string `json:"full_name,omitempty"`

	// Placeholder for description of property is_superuser of obj type User field type str  type boolean
	IsSuperuser *bool `json:"is_superuser,omitempty"`

	// Placeholder for description of property local of obj type User field type str  type boolean
	Local *bool `json:"local,omitempty"`

	// Name of the object.
	// Required: true
	Name *string `json:"name"`

	// password of User.
	Password *string `json:"password,omitempty"`

	// url
	// Read Only: true
	URL *string `json:"url,omitempty"`

	//  It is a reference to an object of type UserAccountProfile.
	UserProfileRef *string `json:"user_profile_ref,omitempty"`

	// username of User.
	Username *string `json:"username,omitempty"`

	// Unique object identifier of the object.
	UUID *string `json:"uuid,omitempty"`
}
