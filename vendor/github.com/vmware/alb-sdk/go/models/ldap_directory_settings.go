// Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
// SPDX-License-Identifier: Apache License 2.0
package models

// This file is auto-generated.

// LdapDirectorySettings ldap directory settings
// swagger:model LdapDirectorySettings
type LdapDirectorySettings struct {

	// LDAP Admin User DN. Administrator credentials are required to search for users under user search DN or groups under group search DN. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	AdminBindDn *string `json:"admin_bind_dn"`

	// Group filter is used to identify groups during search. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GroupFilter *string `json:"group_filter,omitempty"`

	// LDAP group attribute that identifies each of the group members. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GroupMemberAttribute *string `json:"group_member_attribute,omitempty"`

	// Group member entries contain full DNs instead of just user id attribute values. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GroupMemberIsFullDn *bool `json:"group_member_is_full_dn,omitempty"`

	// LDAP group search DN is the root of search for a given group in the LDAP directory. Only matching groups present in this LDAP directory sub-tree will be checked for user membership. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GroupSearchDn *string `json:"group_search_dn,omitempty"`

	// LDAP group search scope defines how deep to search for the group starting from the group search DN. Enum options - AUTH_LDAP_SCOPE_BASE, AUTH_LDAP_SCOPE_ONE, AUTH_LDAP_SCOPE_SUBTREE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	GroupSearchScope *string `json:"group_search_scope,omitempty"`

	// During user or group search, ignore searching referrals. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	IgnoreReferrals *bool `json:"ignore_referrals,omitempty"`

	// LDAP Admin User Password. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	Password *string `json:"password"`

	// LDAP user attributes to fetch on a successful user bind. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserAttributes []string `json:"user_attributes,omitempty"`

	// LDAP user id attribute is the login attribute that uniquely identifies a single user record. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	// Required: true
	UserIDAttribute *string `json:"user_id_attribute"`

	// LDAP user search DN is the root of search for a given user in the LDAP directory. Only user records present in this LDAP directory sub-tree will be validated. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserSearchDn *string `json:"user_search_dn,omitempty"`

	// LDAP user search scope defines how deep to search for the user starting from user search DN. Enum options - AUTH_LDAP_SCOPE_BASE, AUTH_LDAP_SCOPE_ONE, AUTH_LDAP_SCOPE_SUBTREE. Allowed in Enterprise edition with any value, Essentials, Basic, Enterprise with Cloud Services edition.
	UserSearchScope *string `json:"user_search_scope,omitempty"`
}
