package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// UserRole user role
// swagger:model UserRole
type UserRole struct {

	// Placeholder for description of property all_tenants of obj type UserRole field type str  type boolean
	AllTenants *bool `json:"all_tenants,omitempty"`

	// Reference to the Object Access Policy that defines a User's access in the corresponding Tenant. It is a reference to an object of type ObjectAccessPolicy. Field deprecated in 20.1.2. Field introduced in 18.2.7, 20.1.1.
	ObjectAccessPolicyRef *string `json:"object_access_policy_ref,omitempty"`

	//  It is a reference to an object of type Role.
	RoleRef *string `json:"role_ref,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRef *string `json:"tenant_ref,omitempty"`
}
