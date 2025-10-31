// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// AuthMappingRule auth mapping rule
// swagger:model AuthMappingRule
type AuthMappingRule struct {

	// Assignment rule for the Object Access Policy. Enum options - ASSIGN_ALL, ASSIGN_FROM_SELECT_LIST, ASSIGN_MATCHING_GROUP_NAME, ASSIGN_MATCHING_ATTRIBUTE_VALUE, ASSIGN_MATCHING_GROUP_REGEX, ASSIGN_MATCHING_ATTRIBUTE_REGEX, ASSIGN_CONFIG_CONTAINS_ATTRIBUTE_VALUE. Field introduced in 18.2.7, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AssignPolicy *string `json:"assign_policy,omitempty"`

	//  Enum options - ASSIGN_ALL, ASSIGN_FROM_SELECT_LIST, ASSIGN_MATCHING_GROUP_NAME, ASSIGN_MATCHING_ATTRIBUTE_VALUE, ASSIGN_MATCHING_GROUP_REGEX, ASSIGN_MATCHING_ATTRIBUTE_REGEX, ASSIGN_CONFIG_CONTAINS_ATTRIBUTE_VALUE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AssignRole *string `json:"assign_role,omitempty"`

	//  Enum options - ASSIGN_ALL, ASSIGN_FROM_SELECT_LIST, ASSIGN_MATCHING_GROUP_NAME, ASSIGN_MATCHING_ATTRIBUTE_VALUE, ASSIGN_MATCHING_GROUP_REGEX, ASSIGN_MATCHING_ATTRIBUTE_REGEX, ASSIGN_CONFIG_CONTAINS_ATTRIBUTE_VALUE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AssignTenant *string `json:"assign_tenant,omitempty"`

	// Assignment rule for the User Account Profile. Enum options - ASSIGN_ALL, ASSIGN_FROM_SELECT_LIST, ASSIGN_MATCHING_GROUP_NAME, ASSIGN_MATCHING_ATTRIBUTE_VALUE, ASSIGN_MATCHING_GROUP_REGEX, ASSIGN_MATCHING_ATTRIBUTE_REGEX, ASSIGN_CONFIG_CONTAINS_ATTRIBUTE_VALUE. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	AssignUserprofile *string `json:"assign_userprofile,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	AttributeMatch *AuthMatchAttribute `json:"attribute_match,omitempty"`

	// Default tenant ref to assign to user. It is a reference to an object of type Tenant. Field introduced in 22.1.1. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	DefaultTenantRef *string `json:"default_tenant_ref,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GroupMatch *AuthMatchGroupMembership `json:"group_match,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Index *int32 `json:"index"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IsSuperuser *bool `json:"is_superuser,omitempty"`

	// Attribute name for Object Access Policy assignment. Field introduced in 18.2.7, 20.1.1. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	PolicyAttributeName *string `json:"policy_attribute_name,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RoleAttributeName *string `json:"role_attribute_name,omitempty"`

	//  It is a reference to an object of type Role. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	RoleRefs []string `json:"role_refs,omitempty"`

	//  Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantAttributeName *string `json:"tenant_attribute_name,omitempty"`

	//  It is a reference to an object of type Tenant. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	TenantRefs []string `json:"tenant_refs,omitempty"`

	// Attribute name for User Account Profile assignment. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UserprofileAttributeName *string `json:"userprofile_attribute_name,omitempty"`

	// User Account Profile to assign to user on successful match. It is a reference to an object of type UserAccountProfile. Field introduced in 20.1.3. Allowed in Enterprise edition with any value, Enterprise with Cloud Services edition.
	UserprofileRef *string `json:"userprofile_ref,omitempty"`
}
