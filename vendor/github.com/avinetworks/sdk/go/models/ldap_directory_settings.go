package models

// This file is auto-generated.
// Please contact avi-sdk@avinetworks.com for any change requests.

// LdapDirectorySettings ldap directory settings
// swagger:model LdapDirectorySettings
type LdapDirectorySettings struct {

	// LDAP Admin User DN. Administrator credentials are required to search for users under user search DN or groups under group search DN.
	AdminBindDn *string `json:"admin_bind_dn,omitempty"`

	// Group filter is used to identify groups during search.
	GroupFilter *string `json:"group_filter,omitempty"`

	// LDAP group attribute that identifies each of the group members.
	GroupMemberAttribute *string `json:"group_member_attribute,omitempty"`

	// Group member entries contain full DNs instead of just user id attribute values.
	GroupMemberIsFullDn *bool `json:"group_member_is_full_dn,omitempty"`

	// LDAP group search DN is the root of search for a given group in the LDAP directory. Only matching groups present in this LDAP directory sub-tree will be checked for user membership.
	GroupSearchDn *string `json:"group_search_dn,omitempty"`

	// LDAP group search scope defines how deep to search for the group starting from the group search DN. Enum options - AUTH_LDAP_SCOPE_BASE, AUTH_LDAP_SCOPE_ONE, AUTH_LDAP_SCOPE_SUBTREE.
	GroupSearchScope *string `json:"group_search_scope,omitempty"`

	// During user or group search, ignore searching referrals.
	IgnoreReferrals *bool `json:"ignore_referrals,omitempty"`

	// LDAP Admin User Password.
	Password *string `json:"password,omitempty"`

	// LDAP user attributes to fetch on a successful user bind.
	UserAttributes []string `json:"user_attributes,omitempty"`

	// LDAP user id attribute is the login attribute that uniquely identifies a single user record.
	UserIDAttribute *string `json:"user_id_attribute,omitempty"`

	// LDAP user search DN is the root of search for a given user in the LDAP directory. Only user records present in this LDAP directory sub-tree will be validated.
	UserSearchDn *string `json:"user_search_dn,omitempty"`

	// LDAP user search scope defines how deep to search for the user starting from user search DN. Enum options - AUTH_LDAP_SCOPE_BASE, AUTH_LDAP_SCOPE_ONE, AUTH_LDAP_SCOPE_SUBTREE.
	UserSearchScope *string `json:"user_search_scope,omitempty"`
}
