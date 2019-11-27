package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// AuthMappingRule auth mapping rule
// swagger:model AuthMappingRule
type AuthMappingRule struct {

	//  Enum options - ASSIGN_ALL, ASSIGN_FROM_SELECT_LIST, ASSIGN_MATCHING_GROUP_NAME, ASSIGN_MATCHING_ATTRIBUTE_VALUE, ASSIGN_MATCHING_GROUP_REGEX, ASSIGN_MATCHING_ATTRIBUTE_REGEX.
	AssignRole *string `json:"assign_role,omitempty"`

	//  Enum options - ASSIGN_ALL, ASSIGN_FROM_SELECT_LIST, ASSIGN_MATCHING_GROUP_NAME, ASSIGN_MATCHING_ATTRIBUTE_VALUE, ASSIGN_MATCHING_GROUP_REGEX, ASSIGN_MATCHING_ATTRIBUTE_REGEX.
	AssignTenant *string `json:"assign_tenant,omitempty"`

	// Placeholder for description of property attribute_match of obj type AuthMappingRule field type str  type object
	AttributeMatch *AuthMatchAttribute `json:"attribute_match,omitempty"`

	// Placeholder for description of property group_match of obj type AuthMappingRule field type str  type object
	GroupMatch *AuthMatchGroupMembership `json:"group_match,omitempty"`

	// Number of index.
	// Required: true
	Index *int32 `json:"index"`

	// Placeholder for description of property is_superuser of obj type AuthMappingRule field type str  type boolean
	IsSuperuser *bool `json:"is_superuser,omitempty"`

	// role_attribute_name of AuthMappingRule.
	RoleAttributeName *string `json:"role_attribute_name,omitempty"`

	//  It is a reference to an object of type Role.
	RoleRefs []string `json:"role_refs,omitempty"`

	// tenant_attribute_name of AuthMappingRule.
	TenantAttributeName *string `json:"tenant_attribute_name,omitempty"`

	//  It is a reference to an object of type Tenant.
	TenantRefs []string `json:"tenant_refs,omitempty"`
}
